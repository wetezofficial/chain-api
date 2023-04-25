package serviceInterface

import (
	"context"
	"starnet/chain-api/pkg/request"

	"starnet/chain-api/pkg/response"
	"starnet/starnet/models"
)

type IpfsService interface {
	Add(ctx context.Context, apiKey string, addParam request.AddParam, fileList []response.AddResp) error
	ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error
	GetUserIpfsFile(ctx context.Context, apiKey, cid string) (models.IPFSFile, error)
	PinObject(ctx context.Context, apiKey, cid string) error
	UnPinObject(ctx context.Context, apiKey, cid string) error
	CheckMethod(pathStr string) bool
	CheckUserCid(ctx context.Context, apiKey, cid string) bool
	GetApiKeyByActiveGateway(ctx context.Context, subdomain string) string
	GetIpfsUserNoCache(ctx context.Context, apiKey string) (*models.IPFSUser, error)
	IncrIPFSUsage(ctx context.Context, apiKey, setKey string, chainID uint8, addVal int64)
}
