package uploadprovider

import (
	"context"
	"go-simple-api/common"
)

type UploadProvider interface {
	SaveFileUploaded(ctx context.Context, data []byte, dst string) (*common.Image, error)
}