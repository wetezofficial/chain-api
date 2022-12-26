package serviceInterface

import (
	"context"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"mime/multipart"
	"starnet/starnet/models"
)

type IpfsService interface {
	Upload(ctx context.Context, apiKey string, f *multipart.FileHeader) (string, error)
	UploadDir(ctx context.Context, apiKey string, multiFileR *files.MultiFileReader) (string, error)
	GetObject(cidStr string) (io.ReadCloser, error)
	Pin(ctx context.Context, cidStr string) error
	UnPin(ctx context.Context, cidStr string) error
	ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error
}
