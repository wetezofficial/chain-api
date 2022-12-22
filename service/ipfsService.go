package service

import (
	"context"
	"fmt"
	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"starnet/starnet/models"

	serviceInterface "starnet/chain-api/service/interface"
	"starnet/portal-api/pkg/cache"
	daoInterface "starnet/starnet/dao/interface"
)

var _ serviceInterface.IpfsService = &IpfsService{}

// NewIpfsService .
func NewIpfsService(ipfsDao daoInterface.IPFSDao, cache cache.Cache, client client.Client) *IpfsService {
	return &IpfsService{
		ipfsDao:   ipfsDao,
		cache:     cache,
		client:    client,
		ipfsShell: client.IPFS(context.Background()),
	}
}

// IpfsService .
type IpfsService struct {
	ipfsDao   daoInterface.IPFSDao
	cache     cache.Cache
	client    client.Client
	ipfsShell *shell.Shell
}

// UploadDir directory to ipfs cluster
func (s *IpfsService) UploadDir(ctx context.Context, multiFileR *files.MultiFileReader) (string, error) {
	out := make(chan api.AddedOutput)
	e := make(chan error)
	go func() {
		e <- s.client.AddMultiFile(ctx, multiFileR, api.AddParams{}, out)
	}()
	for {
		select {
		case err := <-e:
			return "", err
		case result := <-out:
			return result.Cid.String(), nil
		}
	}
}

// Upload file to ipfs cluster
func (s *IpfsService) Upload(ctx context.Context, r io.Reader) (string, error) {
	return s.ipfsShell.Add(r)
}

func (s *IpfsService) Pin(ctx context.Context, cidStr string) error {
	cid, err := api.DecodeCid(cidStr)
	if err != nil {
		return err
	}
	result, err := s.client.Pin(ctx, cid, api.PinOptions{})
	if err != nil {
		return err
	}
	fmt.Println(result.Cid.String())
	return nil
}

func (s *IpfsService) UnPin(ctx context.Context, cidStr string) error {
	// TODO: check user permission
	cid, err := api.DecodeCid(cidStr)
	if err != nil {
		return err
	}
	result, err := s.client.Unpin(ctx, cid)
	if err != nil {
		return err
	}
	fmt.Println(result.Cid.String())
	return nil
}

func (s *IpfsService) ListUserFile(ctx context.Context, userID int, files *[]models.IPFSFile) error {
	// TODO: get file list form database
	// TODO: portal-api should have a api can list user file by userid
	// TODO: apikey to query userID
	return s.cache.CacheFn(ctx,
		cache.KeyMeta{},
		nil,
		func() error {
			return nil
		},
	)
}
