package serviceInterface

import (
	"context"
	files "github.com/ipfs/go-ipfs-files"
	"mime/multipart"
	"starnet/starnet/models"
)

type IpfsService interface {
	Upload(ctx context.Context, apiKey string, f *multipart.FileHeader) (string, error)
	UploadDir(ctx context.Context, apiKey string, multiFileR *files.MultiFileReader) (string, error)
	Pin(ctx context.Context, cidStr string) error
	ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error
}
