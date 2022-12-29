package serviceInterface

import (
	"context"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"
	"starnet/starnet/models"
)

type IpfsService interface {
	Add(ctx context.Context, apiKey string, apiParam request.AddParam, multiFileR *files.MultiFileReader) ([]response.AddResp, error)
	GetObject(cidStr string) (io.ReadCloser, error)
	Pin(ctx context.Context, cidStr string) error
	UnPin(ctx context.Context, apiKey, cidStr string) error
	ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error
}
