package service

import (
	"context"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"

	"starnet/portal-api/pkg/cache"
	serviceInterface "starnet/portal-api/service/interface"
	daoInterface "starnet/starnet/dao/interface"
)

var _ serviceInterface.IpfsService = &IpfsService{}

// NewIpfsService TODO: add ipfs client
func NewIpfsService(ipfsDao daoInterface.IPFSDao, cache cache.Cache) *IpfsService {
	return &IpfsService{ipfsDao: ipfsDao, cache: cache}
}

// IpfsService TODO: add ipfs client
type IpfsService struct {
	ipfsDao daoInterface.IPFSDao
	cache   cache.Cache
}

func (s *IpfsService) Upload(ctx context.Context, c client.Client) error {
	//TODO implement me
	panic("implement me")
}

func (s *IpfsService) Pin(ctx context.Context, c client.Client) error {
	//TODO implement me
	panic("implement me")
}

func (s *IpfsService) GetFile(ctx context.Context, c client.Client) interface{} {
	//TODO implement me
	panic("implement me")
}

func (s *IpfsService) ListUserCid(ctx context.Context, c client.Client) error {
	return s.cache.CacheFn(ctx,
		cache.KeyMeta{},
		nil,
		func() error {
			return nil
		},
	)
}
