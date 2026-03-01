package pdf

import (
	"bytes"
	"fmt"
	"strings"

	gopdf "github.com/ledongthuc/pdf"
)

type PdfCpuExtractor struct{}

func NewPdfCpuExtractor() *PdfCpuExtractor {
	return &PdfCpuExtractor{}
}

func (e *PdfCpuExtractor) Extract(filePath string) (string, error) {
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
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("error extracting page %d: %w", i, err)
		}
		buf.WriteString(text)
	}

	return strings.TrimSpace(buf.String()), nil
}
