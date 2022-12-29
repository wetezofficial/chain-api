package service

import (
	"context"
	"fmt"
	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"
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

// Add file or directory to ipfs cluster
func (s *IpfsService) Add(ctx context.Context, apiKey string, apiParam request.AddParam, multiFileR *files.MultiFileReader) ([]response.AddResp, error) {
	out := make(chan api.AddedOutput)
	e := make(chan error)
	userID := s.getUserIDByAPIKey(ctx, apiKey)
	go func() {
		e <- s.client.AddMultiFile(ctx, multiFileR, api.AddParams{
			Recursive: apiParam.Recursive,
			Hidden:    apiParam.Hidden,
			Wrap:      apiParam.WrapWithDirectory,
			NoPin:     !apiParam.Pin,
			IPFSAddParams: api.IPFSAddParams{
				Chunker:    apiParam.Chunker,
				Progress:   apiParam.Progress,
				CidVersion: apiParam.CidVersion,
				NoCopy:     apiParam.Nocopy,
			},
		}, out)
	}()
	var dirHash string
	var results []response.AddResp
	var dbFiles []*models.IPFSFile
	pinStatus := models.PinStatusUnPin
	if apiParam.Pin {
		pinStatus = models.PinStatusPin
	}
	for {
		select {
		case err := <-e:
			if err != nil {
				return nil, err
			}
			goto SAVED
		case result := <-out:
			if result.Cid.String() != "b" {
				hash := result.Cid.String()
				if result.Name == "" {
					dirHash = hash
				}
				results = append(results, response.AddResp{
					Name: result.Name,
					Hash: hash,
					Size: result.Size,
				})
				dbFiles = append(dbFiles, &models.IPFSFile{
					PinStatus: pinStatus,
					UserId:    userID,
					FileSize:  result.Size,
					FileName:  result.Name,
					CID:       hash,
				})
			}
		}
	}
SAVED:
	if dirHash != "" {
		for k := range dbFiles {
			dbFiles[k].WrapWithDirCid = dirHash
			dbFiles[k].WrapDirName = apiParam.WrapDirName
		}
	}
	if err := s.ipfsDao.BatchSaveFiles(dbFiles); err != nil {
		return nil, err
	}
	return results, nil
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

func (s *IpfsService) UnPin(ctx context.Context, apiKey, cidStr string) error {
	_, err := s.getUserFile(ctx, apiKey, cidStr)
	if err != nil {
		return err
	}
	cid, err := api.DecodeCid(cidStr)
	if err != nil {
		return err
	}
	_, err = s.client.Unpin(ctx, cid)
	if err != nil {
		return err
	}
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

func (s *IpfsService) getUserFile(ctx context.Context, apiKey, cid string) (*models.IPFSFile, error) {
	file := new(models.IPFSFile)
	if err := s.cache.CacheFn(ctx,
		cachekey.UserIPFSFiles(apiKey),
		file,
		func() error {
			return s.ipfsDao.GetUserFile(s.getUserIDByAPIKey(ctx, apiKey), cid, file)
		},
	); err != nil {
		return nil, err
	}
	return file, nil
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
