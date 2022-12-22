package serviceInterface

import (
	"context"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"starnet/starnet/models"
)

type IpfsService interface {
	Upload(ctx context.Context, r io.Reader) (string, error)
	UploadDir(ctx context.Context, multiFileR *files.MultiFileReader) (string, error)
	Pin(ctx context.Context, cidStr string) error
	ListUserFile(ctx context.Context, userID int, files *[]models.IPFSFile) error
}
