package service

import (
	"context"
	"fmt"
	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"mime/multipart"
	serviceInterface "starnet/chain-api/service/interface"
	"starnet/portal-api/app/cachekey"
	"starnet/portal-api/pkg/cache"
	daoInterface "starnet/starnet/dao/interface"
	"starnet/starnet/models"
)

var _ serviceInterface.IpfsService = &IpfsService{}

// NewIpfsService .
func NewIpfsService(ipfsDao daoInterface.IPFSDao, userDao daoInterface.UserDao, cache cache.Cache, client client.Client) *IpfsService {
	return &IpfsService{
		ipfsDao:   ipfsDao,
		userDao:   userDao,
		cache:     cache,
		client:    client,
		ipfsShell: client.IPFS(context.Background()),
	}
}

// IpfsService .
type IpfsService struct {
	ipfsDao   daoInterface.IPFSDao
	userDao   daoInterface.UserDao
	cache     cache.Cache
	client    client.Client
	ipfsShell *shell.Shell
}

// UploadDir directory to ipfs cluster
func (s *IpfsService) UploadDir(ctx context.Context, apiKey string, multiFileR *files.MultiFileReader) (string, error) {
	out := make(chan api.AddedOutput)
	e := make(chan error)
	go func() {
		e <- s.client.AddMultiFile(ctx, multiFileR, api.AddParams{}, out)
	}()
	var rootCid string
	for {
		select {
		case err := <-e:
			if err != nil {
				return "", err
			}
			goto SAVED
		case result := <-out:
			if result.Cid.String() != "b" {
				rootCid = result.Cid.String()
			}
		}
	}
SAVED:
	if err := s.ipfsDao.SaveFile(&models.IPFSFile{
		UserId:    s.getUserIDByAPIKey(ctx, apiKey),
		CID:       rootCid,
		PinStatus: models.PinStatusUnPin,
		// TODO:
		FileName: "",
		FileSize: 0,
	}); err != nil {
		return "", err
	}
	return rootCid, nil
}

// Upload file to ipfs cluster
func (s *IpfsService) Upload(ctx context.Context, apiKey string, f *multipart.FileHeader) (string, error) {
	file, err := f.Open()
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)
	hash, err := s.ipfsShell.Add(io.Reader(file))
	if err != nil {
		return "", err
	}
	if err := s.ipfsDao.SaveFile(&models.IPFSFile{
		UserId:    s.getUserIDByAPIKey(ctx, apiKey),
		CID:       hash,
		PinStatus: models.PinStatusUnPin,
		FileName:  f.Filename,
		FileSize:  f.Size,
	}); err != nil {
		return "", err
	}
	return hash, nil
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

func (s *IpfsService) GetObject(cidStr string) (io.ReadCloser, error) {
	return s.ipfsShell.Cat(cidStr)
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

func (s *IpfsService) ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile) error {
	return s.cache.CacheFn(ctx,
		cachekey.UserIPFSFiles(apiKey),
		files,
		func() error {
			return s.ipfsDao.ListUserFile(s.getUserIDByAPIKey(ctx, apiKey), files)
		},
	)
}

func (s *IpfsService) getUserIDByAPIKey(ctx context.Context, apiKey string) int {
	var userID int
	_ = s.cache.CacheFn(ctx,
		cachekey.APIKeyUserID(apiKey),
		&userID,
		func() error {
			var err error
			userID, err = s.userDao.GetIDByAPIKey(apiKey)
			return err
		},
	)
	return userID
}
