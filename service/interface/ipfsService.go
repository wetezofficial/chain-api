package serviceInterface

import (
	"context"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
)

type IpfsService interface {
	ListUserCid(ctx context.Context, c client.Client) error
	Upload(ctx context.Context, c client.Client) error
	Pin(ctx context.Context, c client.Client) error
	GetFile(ctx context.Context, c client.Client) interface{}
}
