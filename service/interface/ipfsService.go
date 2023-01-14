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
	Add(ctx context.Context, apiKey string, fileList []response.AddResp) error
	AddCluster(ctx context.Context, apiKey string, apiParam request.AddParam, multiFileR *files.MultiFileReader) ([]response.AddResp, error)
	GetObject(cidStr string) (io.ReadCloser, error)
	ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error
	CheckMethod(pathStr string) bool
	CheckUserCid(ctx context.Context, apiKey, cid string) bool
	GetIpfsUserNoCache(ctx context.Context, apiKey string) (*models.IPFSUser, error)
	UpdateIPFSUsage(ctx context.Context, apiKey, setKey string, chainID uint8, newVal interface{})
}
