package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"contract-scanner/internal/usecase/providers"

	gopdf "github.com/ledongthuc/pdf"
)

const (
	sourceNative = "native"
	sourceOCR    = "ocr"
	sourceHybrid = "hybrid"
)

type Config struct {
	OCREnabled      bool
	OCRLang         string
	OCRDPI          int
	MinQualityScore float64
	OCRMaxPages     int
	OCRTimeout      time.Duration
}

type HybridExtractor struct {
	config         Config
	nativeExtract  func(filePath string) (string, error)
	ocrExtractWith func(ctx context.Context, filePath string) (string, error)
}

type cleaningStats struct {
	inputLines       int
	outputLines      int
	removedWatermark int
	removedDuplicate int
	hadControlChars  bool
}

func DefaultConfig() Config {
	return Config{
		OCREnabled:      true,
		OCRLang:         "por+eng",
		OCRDPI:          300,
		MinQualityScore: 0.35,
		OCRMaxPages:     80,
		OCRTimeout:      120 * time.Second,
	}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if v := strings.TrimSpace(os.Getenv("PDF_OCR_ENABLED")); v != "" {
		cfg.OCREnabled = strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes")
	}
	if v := strings.TrimSpace(os.Getenv("PDF_OCR_LANG")); v != "" {
		cfg.OCRLang = v
	}
	if v := strings.TrimSpace(os.Getenv("PDF_OCR_DPI")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.OCRDPI = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("PDF_MIN_QUALITY_SCORE")); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.MinQualityScore = clamp(parsed)
		}
	}
	if v := strings.TrimSpace(os.Getenv("PDF_OCR_MAX_PAGES")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.OCRMaxPages = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("PDF_OCR_TIMEOUT_SEC")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.OCRTimeout = time.Duration(parsed) * time.Second
		}
	}

	return cfg
}

func NewHybridExtractor(config Config) *HybridExtractor {
	extractor := &HybridExtractor{config: config}
	extractor.nativeExtract = extractNative
	extractor.ocrExtractWith = extractor.extractWithOCR
	return extractor
}

func NewPdfCpuExtractor() *HybridExtractor {
	return NewHybridExtractor(ConfigFromEnv())
}

func (e *HybridExtractor) Extract(ctx context.Context, filePath string) (*providers.ExtractResult, error) {
	log.Printf("[DEBUG][pdf-extractor] start extract file=%s", filePath)
	warnings := []string{}

	nativeRaw, nativeErr := e.nativeExtract(filePath)
	if nativeErr != nil {
		log.Printf("[DEBUG][pdf-extractor] native extract failed: %v", nativeErr)
		warnings = append(warnings, "native extraction failed")
	}
	nativeText, nativeStats := cleanAndNormalize(nativeRaw)
	nativeScore := scoreQuality(nativeText)
	nativeUseful := isMinimallyUseful(nativeText)
	log.Printf(
		"[DEBUG][pdf-extractor] native done score=%.2f useful=%t stats={in:%d out:%d watermark:%d dup:%d ctrl:%t}",
		nativeScore,
		nativeUseful,
		nativeStats.inputLines,
		nativeStats.outputLines,
		nativeStats.removedWatermark,
		nativeStats.removedDuplicate,
		nativeStats.hadControlChars,
	)
	warnings = append(warnings, statsWarnings(nativeStats)...)

	if nativeErr == nil && nativeScore >= e.config.MinQualityScore {
		log.Printf("[DEBUG][pdf-extractor] selecting native result score=%.2f threshold=%.2f", nativeScore, e.config.MinQualityScore)
		return &providers.ExtractResult{
			Text:         nativeText,
			Source:       sourceNative,
			QualityScore: nativeScore,
			Warnings:     dedupeWarnings(warnings),
		}, nil
	}

	if !e.config.OCREnabled {
		log.Printf("[DEBUG][pdf-extractor] OCR disabled; native below threshold score=%.2f threshold=%.2f", nativeScore, e.config.MinQualityScore)
		if nativeUseful {
			if nativeScore < e.config.MinQualityScore {
				warnings = append(warnings, "text extraction quality is low; analysis may be partial")
			}
			return &providers.ExtractResult{
				Text:         nativeText,
				Source:       sourceNative,
				QualityScore: nativeScore,
				Warnings:     dedupeWarnings(warnings),
			}, nil
		}
		return nil, fmt.Errorf("unable to extract readable contract text")
	}

	log.Printf(
		"[DEBUG][pdf-extractor] start OCR fallback score=%.2f threshold=%.2f cfg={lang:%s dpi:%d max_pages:%d timeout:%s}",
		nativeScore,
		e.config.MinQualityScore,
		e.config.OCRLang,
		e.config.OCRDPI,
		e.config.OCRMaxPages,
		e.config.OCRTimeout.String(),
	)
	ocrRaw, ocrErr := e.ocrExtractWith(ctx, filePath)
	if ocrErr != nil {
		log.Printf("[DEBUG][pdf-extractor] OCR fallback failed: %v", ocrErr)
		warnings = append(warnings, fmt.Sprintf("ocr fallback failed: %v", ocrErr))
	}
	ocrText, ocrStats := cleanAndNormalize(ocrRaw)
	ocrScore := scoreQuality(ocrText)
	ocrUseful := isMinimallyUseful(ocrText)
	log.Printf(
		"[DEBUG][pdf-extractor] OCR done score=%.2f useful=%t stats={in:%d out:%d watermark:%d dup:%d ctrl:%t}",
		ocrScore,
		ocrUseful,
		ocrStats.inputLines,
		ocrStats.outputLines,
		ocrStats.removedWatermark,
		ocrStats.removedDuplicate,
		ocrStats.hadControlChars,
	)
	warnings = append(warnings, statsWarnings(ocrStats)...)

	switch {
	case nativeUseful && ocrUseful:
		chosenText := nativeText
		chosenScore := nativeScore
		if ocrScore > nativeScore {
			chosenText = ocrText
			chosenScore = ocrScore
		}
		if chosenScore < e.config.MinQualityScore {
			warnings = append(warnings, "text extraction quality is low; analysis may be partial")
		}
		log.Printf("[DEBUG][pdf-extractor] selecting hybrid result score=%.2f", chosenScore)
		return &providers.ExtractResult{
			Text:         chosenText,
			Source:       sourceHybrid,
			QualityScore: chosenScore,
			Warnings:     dedupeWarnings(warnings),
		}, nil
	case ocrUseful:
		if ocrScore < e.config.MinQualityScore {
			warnings = append(warnings, "text extraction quality is low; analysis may be partial")
		}
		log.Printf("[DEBUG][pdf-extractor] selecting OCR result score=%.2f", ocrScore)
		return &providers.ExtractResult{
			Text:         ocrText,
			Source:       sourceOCR,
			QualityScore: ocrScore,
			Warnings:     dedupeWarnings(warnings),
		}, nil
	case nativeUseful:
		warnings = append(warnings, "using native extraction due to OCR limitations")
		if nativeScore < e.config.MinQualityScore {
			warnings = append(warnings, "text extraction quality is low; analysis may be partial")
		}
		log.Printf("[DEBUG][pdf-extractor] selecting native result after OCR attempt score=%.2f", nativeScore)
		return &providers.ExtractResult{
			Text:         nativeText,
			Source:       sourceNative,
			QualityScore: nativeScore,
			Warnings:     dedupeWarnings(warnings),
		}, nil
	default:
		log.Printf("[DEBUG][pdf-extractor] no useful text from native or OCR")
		return nil, fmt.Errorf("unable to extract readable contract text")
	}
}

func (e *HybridExtractor) extractWithOCR(ctx context.Context, filePath string) (string, error) {
	log.Printf("[DEBUG][pdf-extractor] OCR pipeline started file=%s", filePath)
	ocrCtx := ctx
	cancel := func() {}
	if e.config.OCRTimeout > 0 {
		log.Printf("[DEBUG][pdf-extractor] applying OCR timeout=%s", e.config.OCRTimeout.String())
		ocrCtx, cancel = context.WithTimeout(ctx, e.config.OCRTimeout)
	}
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "contract-ocr-*")
	if err != nil {
		return "", fmt.Errorf("creating ocr temp dir: %w", err)
	}
	log.Printf("[DEBUG][pdf-extractor] temp dir created: %s", tmpDir)
	defer os.RemoveAll(tmpDir)

	prefix := filepath.Join(tmpDir, "page")
	args := []string{"-r", strconv.Itoa(e.config.OCRDPI), "-png", filePath, prefix}
	if e.config.OCRMaxPages > 0 {
		args = append([]string{"-f", "1", "-l", strconv.Itoa(e.config.OCRMaxPages)}, args...)
	}
	log.Printf("[DEBUG][pdf-extractor] running pdftoppm args=%v", args)

	if out, err := exec.CommandContext(ocrCtx, "pdftoppm", args...).CombinedOutput(); err != nil {
		if errors.Is(ocrCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("ocr conversion timed out")
		}
		return "", fmt.Errorf("running pdftoppm: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	log.Printf("[DEBUG][pdf-extractor] pdftoppm finished successfully")

	images, err := filepath.Glob(filepath.Join(tmpDir, "page-*.png"))
	if err != nil {
		return "", fmt.Errorf("listing generated images: %w", err)
	}
	if len(images) == 0 {
		return "", fmt.Errorf("no images generated for ocr")
	}
	sort.Strings(images)
	log.Printf("[DEBUG][pdf-extractor] generated %d page images", len(images))

	var buf strings.Builder
	for idx, image := range images {
		log.Printf("[DEBUG][pdf-extractor] running tesseract page=%d/%d image=%s lang=%s", idx+1, len(images), filepath.Base(image), e.config.OCRLang)
		out, err := exec.CommandContext(ocrCtx, "tesseract", image, "stdout", "-l", e.config.OCRLang, "--psm", "6").CombinedOutput()
		if err != nil {
			if errors.Is(ocrCtx.Err(), context.DeadlineExceeded) {
				return "", fmt.Errorf("ocr processing timed out")
			}
			return "", fmt.Errorf("running tesseract on %s: %w (%s)", filepath.Base(image), err, strings.TrimSpace(string(out)))
		}
		log.Printf("[DEBUG][pdf-extractor] tesseract page=%d done bytes=%d", idx+1, len(out))
		buf.WriteString(string(out))
		buf.WriteString("\n")
	}

	log.Printf("[DEBUG][pdf-extractor] OCR pipeline finished total_bytes=%d", buf.Len())
	return strings.TrimSpace(buf.String()), nil
}

func extractNative(filePath string) (string, error) {
	f, reader, err := gopdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening pdf: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, pageErr := page.GetPlainText(nil)
		if pageErr != nil {
			return "", fmt.Errorf("error extracting page %d: %w", i, pageErr)
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return strings.TrimSpace(buf.String()), nil
}

func cleanAndNormalize(raw string) (string, cleaningStats) {
	stats := cleaningStats{}
	if strings.TrimSpace(raw) == "" {
		return "", stats
	}

	sanitized, removedControl := removeInvalidControls(raw)
	stats.hadControlChars = removedControl > 0

	lines := strings.Split(sanitized, "\n")
	stats.inputLines = len(lines)

	allowedRepeats := map[string]int{}
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = normalizeLine(line)
		if line == "" {
			continue
		}

		if isWatermarkLine(line) {
			stats.removedWatermark++
			continue
		}

		count := allowedRepeats[line]
		if count >= maxRepeatsForLine(line) {
			stats.removedDuplicate++
			continue
		}
		allowedRepeats[line] = count + 1
		filtered = append(filtered, line)
	}

	stats.outputLines = len(filtered)
	return strings.TrimSpace(strings.Join(filtered, "\n")), stats
}

func statsWarnings(stats cleaningStats) []string {
	warnings := []string{}
	if stats.hadControlChars {
		warnings = append(warnings, "extracted text contained control characters and was normalized")
	}
	if stats.removedWatermark > 0 {
		warnings = append(warnings, "signature watermark text was removed during normalization")
	}
	if stats.removedDuplicate > 0 {
		warnings = append(warnings, "repeated lines were collapsed during normalization")
	}
	return warnings
}

func scoreQuality(text string) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	lines := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return 0
	}

	charLen := len([]rune(text))
	lengthScore := clamp(float64(charLen) / 4000)

	lineCount := map[string]int{}
	for _, line := range lines {
		lineCount[line]++
	}
	uniqueRatio := float64(len(lineCount)) / float64(len(lines))

	duplicateLines := 0
	for _, count := range lineCount {
		if count > 1 {
			duplicateLines += count - 1
		}
	}
	duplicateRatio := float64(duplicateLines) / float64(len(lines))
	diversityScore := 1 - duplicateRatio
	if diversityScore < 0 {
		diversityScore = 0
	}

	legalTerms := []string{
		"contrato", "cláusula", "clausula", "partes", "objeto", "vigência", "vigencia",
		"rescisão", "rescisao", "prazo", "multa", "pagamento", "responsabilidade",
	}
	lower := strings.ToLower(text)
	termsFound := 0
	for _, term := range legalTerms {
		if strings.Contains(lower, term) {
			termsFound++
		}
	}
	termScore := clamp(float64(termsFound) / 6)

	score := 0.40*lengthScore + 0.25*uniqueRatio + 0.20*termScore + 0.15*diversityScore
	return clamp(score)
}

func isMinimallyUseful(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	if len([]rune(trimmed)) < 160 {
		return false
	}
	return scoreQuality(trimmed) >= 0.12
}

func isWatermarkLine(line string) bool {
	line = strings.ToLower(line)
	patterns := []string{
		"d4sign",
		"documento assinado eletronicamente",
		"secure.d4sign.com.br",
		"para confirmar as assinaturas acesse",
		"certificado de assinaturas gerado",
		"sincronizado com o ntp.br",
	}
	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	return false
}

func normalizeLine(line string) string {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func maxRepeatsForLine(line string) int {
	runes := []rune(strings.TrimSpace(line))
	if len(runes) <= 40 {
		return 1
	}
	if len(runes) <= 120 {
		return 2
	}
	return 3
}

func removeInvalidControls(input string) (string, int) {
	removed := 0
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			removed++
			return -1
		}
		return r
	}, input)
	return cleaned, removed
}

func dedupeWarnings(warnings []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		warning = strings.TrimSpace(warning)
		if warning == "" {
			continue
		}
		if _, exists := seen[warning]; exists {
			continue
		}
		seen[warning] = struct{}{}
		result = append(result, warning)
	}
	return result
}

func clamp(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
