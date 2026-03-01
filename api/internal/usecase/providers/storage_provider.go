package providers

import "context"

type StorageProvider interface {
	GeneratePresignedPutURL(ctx context.Context, key string, contentType string) (string, error)
	HeadObject(ctx context.Context, key string) error
	GetObject(ctx context.Context, key string, destPath string) error
}
