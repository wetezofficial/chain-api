package service

import (
	"context"
	"github.com/spf13/cast"
	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"

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

// Add file info to the database
func (s *IpfsService) Add(ctx context.Context, apiKey string, pinStatus uint8, fileList []response.AddResp) error {
	userID := s.getUserIDByAPIKey(ctx, apiKey)
	var dbFiles []*models.IPFSFile
	for _, v := range fileList {
		dbFiles = append(dbFiles, &models.IPFSFile{
			PinStatus: pinStatus,
			UserId:    userID,
			FileSize:  cast.ToUint64(v.Size),
			FileName:  v.Name,
			CID:       v.Hash,
		})
	}
	return s.ipfsDao.BatchSaveFiles(dbFiles)
}

func (s *IpfsService) Pin(ctx context.Context, cidStr string, apiParam request.PinParam) error {
	//stat, err := s.ipfsShell.FilesStat(ctx, fileList[len(fileList)-1].Hash)
	//if err != nil {
	//	return err
	//}
	//if stat.Type == "directory" {
	//	dbFiles[len(dbFiles)-1].WrapWithDirCid = fileList[len(fileList)-1].Hash
	//	dbFiles[len(dbFiles)-1].WrapDirName = fileList[len(fileList)-1].Name
	//	list, err := s.ipfsShell.List(fileList[len(fileList)-1].Hash)
	//	if err != nil {
	//		return err
	//	}
	//	for _, v := range list {
	//		if v.Type == 1 {
	//			dbFiles[len(dbFiles)-1].WrapDirName = v.Name
	//			break
	//		}
	//	}
	//}
	return nil
}

func (s *IpfsService) UnPin(ctx context.Context, apiKey, cidStr string) error {
	_, err := s.getUserFile(ctx, apiKey, cidStr)
	if err != nil {
		return err
	}
	// TODO: save to db
	return nil
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
