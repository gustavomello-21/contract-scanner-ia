# Contract Scanner API

## OCR e extração de PDF

A API usa extração híbrida de texto:
- extração nativa de PDF (camada textual)
- fallback OCR com `pdftoppm` + `tesseract`

Isso melhora casos de contratos escaneados, assinados eletronicamente e PDFs com texto corrompido.

## Variáveis de ambiente

Configure no `.env` (raiz do projeto):

```env
PDF_OCR_ENABLED=true
PDF_OCR_LANG=por+eng
PDF_OCR_DPI=300
PDF_MIN_QUALITY_SCORE=0.35
PDF_OCR_MAX_PAGES=80
PDF_OCR_TIMEOUT_SEC=120
```

## Dependências locais

### Linux (Debian/Ubuntu)

```bash
sudo apt-get update
sudo apt-get install -y tesseract-ocr tesseract-ocr-por poppler-utils
```

### macOS (Homebrew)

```bash
brew install tesseract poppler
brew install tesseract-lang
```

### Windows

1. Instale Tesseract OCR (com idioma `por`) e adicione ao `PATH`.
2. Instale Poppler for Windows e adicione ao `PATH` (`pdftoppm.exe`).
3. Reinicie o terminal e valide:

```bash
tesseract --version
pdftoppm -v
```

## Docker

O `api/Dockerfile` já instala as dependências OCR no runtime:
- `tesseract-ocr`
- `tesseract-ocr-data-por`
- `poppler-utils`

## Comportamento da extração

- Se texto nativo tiver qualidade suficiente, usa `source=native`.
- Se estiver ruim, tenta OCR e escolhe o melhor texto (`source=ocr` ou `source=hybrid`).
- Se qualidade final for baixa, a análise continua com avisos em `analysis_warnings`.
- Se nada legível for extraído, a API retorna erro: `unable to extract readable contract text`.
