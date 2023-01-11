package serviceInterface

import (
	"context"
	"io"
	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"
	"starnet/starnet/models"

	files "github.com/ipfs/go-ipfs-files"
)

type IpfsService interface {
	Add(ctx context.Context, apiKey string, apiParam request.AddParam, multiFileR *files.MultiFileReader) ([]response.AddResp, error)
	GetObject(cidStr string) (io.ReadCloser, error)
	Pin(ctx context.Context, cidStr string, apiParam request.PinParam,) error
	UnPin(ctx context.Context, apiKey, cidStr string, apiParam request.PinParam,) error
	ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error
	UpdateUserTotalSave(ctx context.Context,apiKey string, fileSize int64) (int64, error)
}
