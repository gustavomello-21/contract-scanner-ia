package providers

type PDFExtractor interface {
	Extract(filePath string) (string, error)
}
