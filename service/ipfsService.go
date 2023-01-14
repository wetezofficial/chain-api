package service

import (
	"context"

	"starnet/chain-api/pkg/response"

	"github.com/spf13/cast"

	serviceInterface "starnet/chain-api/service/interface"
	"starnet/portal-api/app/cachekey"
	"starnet/portal-api/pkg/cache"
	daoInterface "starnet/starnet/dao/interface"
	"starnet/starnet/models"

	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
)

var _ serviceInterface.IpfsService = &IpfsService{}

// NewIpfsService .
func NewIpfsService(ipfsDao daoInterface.IPFSDao, userDao daoInterface.UserDao, cache cache.Cache, client client.Client) *IpfsService {
	allowMethod := []string{
		"/add",
		"/block/get",
		"/block/put",
		"/block/stat",
		"/cat",
		"/dag/get",
		"/dag/put",
		"/dag/resolve",
		"/get",
		"/pin/add",
		"/pin/ls",
		"/pin/rm",
		"/pin/update",
		"/version",
	}
	return &IpfsService{
		ipfsDao:     ipfsDao,
		userDao:     userDao,
		cache:       cache,
		client:      client,
		ipfsShell:   client.IPFS(context.Background()),
		allowMethod: allowMethod,
	}
}

// IpfsService .
type IpfsService struct {
	ipfsDao     daoInterface.IPFSDao
	userDao     daoInterface.UserDao
	cache       cache.Cache
	client      client.Client
	ipfsShell   *shell.Shell
	allowMethod []string
}

func (s *IpfsService) CheckUserCid(ctx context.Context, apiKey, cid string) bool {
	var result bool
	_ = s.cache.CacheFn(ctx,
		cachekey.CheckUserCID(apiKey, cid),
		&result,
		func() error {
			file := new(models.IPFSFile)
			_ = s.ipfsDao.GetUserFile(s.getUserIDByAPIKey(ctx, apiKey), cid, file)
			result = file.ID > 0
			return nil
		},
	)
	return result
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

func (s *IpfsService) GetIPFSUser(ctx context.Context, apiKey string, files *[]models.IPFSUser) error {
	return s.cache.CacheFn(ctx,
		cachekey.UserIPFSFiles(apiKey),
		files,
		func() error {
			return s.ipfsDao.ListUserFile(s.getUserIDByAPIKey(ctx, apiKey), files)
		},
	)
}

// Add file info to the database
func (s *IpfsService) Add(ctx context.Context, apiKey string, fileList []response.AddResp) error {
	userID := s.getUserIDByAPIKey(ctx, apiKey)
	var dbFiles []*models.IPFSFile
	for _, v := range fileList {
		dbFiles = append(dbFiles, &models.IPFSFile{
			UserId:   userID,
			FileSize: cast.ToUint64(v.Size),
			FileName: v.Name,
			CID:      v.Hash,
		})
	}
	var addStorage uint64
	for _, v := range dbFiles {
		count, _ := s.ipfsDao.CountUserFile(userID, v.CID)
		if count == 1 {
			addStorage += v.FileSize
		}
	}
	if addStorage > 0 {
		// newValue, err := s.ipfsDao.IncrUserStorage(userID, addStorage)
		_, err := s.ipfsDao.IncrUserStorage(userID, addStorage)
		if err == nil {
			// TODO: update redis
			// s.cache.Set(ctx, cachekey.APIKeyUserID(apiKey string), data interface{})
		}
	}
	return s.ipfsDao.BatchSaveFiles(dbFiles)
}

func (s *IpfsService) GetIpfsUser(ctx context.Context, apiKey string) (*models.IPFSUser, error) {
	// TODO: update after add & BandwidthHook ？
	user := new(models.IPFSUser)
	if err := s.cache.CacheFn(ctx,
		cachekey.UserIPFSFiles(apiKey),
		user,
		func() error {
			return nil
			// return s.ipfsDao.GetUserFile(s.getUserIDByAPIKey(ctx, apiKey), "cid", user)
		},
	); err != nil {
		return nil, err
	}
	return user, nil
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
