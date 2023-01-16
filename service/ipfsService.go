package service

import (
	"context"

	"starnet/chain-api/pkg/response"

	"github.com/spf13/cast"
	"go.uber.org/zap"

	serviceInterface "starnet/chain-api/service/interface"
	"starnet/portal-api/app/cachekey"
	commonKey "starnet/starnet/cachekey"
	"starnet/starnet/constant"

	"starnet/portal-api/pkg/cache"
	daoInterface "starnet/starnet/dao/interface"
	"starnet/starnet/models"
)

var _ serviceInterface.IpfsService = &IpfsService{}

// IpfsService .
type IpfsService struct {
	ipfsDao     daoInterface.IPFSDao
	userDao     daoInterface.UserDao
	cache       cache.Cache
	allowMethod []string
	logger      *zap.Logger
}

// NewIpfsService .
func NewIpfsService(ipfsDao daoInterface.IPFSDao, userDao daoInterface.UserDao, cache cache.Cache, logger *zap.Logger) *IpfsService {
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
		allowMethod: allowMethod,
		logger:      logger,
	}
}

func (s *IpfsService) CheckMethod(pathStr string) bool {
	for _, v := range s.allowMethod {
		if v == pathStr {
			return true
		}
	}
	return false
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
		count, err := s.ipfsDao.CountUserFile(userID, v.CID)
		if err != nil {
			s.logger.With(zap.Int("userID", userID), zap.String("cid", v.CID)).Error("count user file failed", zap.Error(err))
		}
		if count == 0 {
			addStorage += v.FileSize
		}
	}
	if addStorage > 0 {
		// newValue, err := s.ipfsDao.IncrUserStorage(userID, addStorage)
		_, err := s.ipfsDao.IncrUserStorage(userID, addStorage)
		if err == nil {
			s.IncrIPFSUsage(
				ctx,
				apiKey,
				commonKey.IpfsLimitStorageSetKey(),
				constant.ChainIPFS.ChainID,
				int64(addStorage),
			)
		}
	}
	return s.ipfsDao.BatchSaveFiles(dbFiles)
}

func (s *IpfsService) IncrIPFSUsage(ctx context.Context, apiKey, setKey string, chainID uint8, addVal int64) {
	logger := s.logger.With(zap.String("apiKey", apiKey), zap.Uint8("chainId", chainID))

	_, err := s.cache.HIncrBy(context.TODO(), cachekey.GetUserIPFSUsageKeyMate(apiKey, chainID).Key, setKey, addVal)
	if err != nil {
		logger.Error("set user auth usage failed", zap.Error(err))
	}
}

// GetIpfsUserNoCache .
func (s *IpfsService) GetIpfsUserNoCache(ctx context.Context, apiKey string) (*models.IPFSUser, error) {
	user := new(models.IPFSUser)
	return user, s.ipfsDao.GetIPFSUser(s.getUserIDByAPIKey(ctx, apiKey), user)
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
