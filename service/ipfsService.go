package service

import (
	"context"
	"errors"
	"starnet/chain-api/pkg/request"

	"starnet/chain-api/pkg/response"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	serviceInterface "starnet/chain-api/service/interface"
	"starnet/starnet/cachekey"
	"starnet/starnet/constant"

	daoInterface "starnet/starnet/dao/interface"
	"starnet/starnet/models"
	"starnet/starnet/pkg/cache"
)

var _ serviceInterface.IpfsService = &IpfsService{}

// IpfsService .
type IpfsService struct {
	ipfsDao     daoInterface.IPFSDao
	userDao     daoInterface.UserDao
	projectDao  daoInterface.ProjectDao
	rdb         redis.UniversalClient
	cache       cache.Cache
	allowMethod []string
	logger      *zap.Logger
}

// NewIpfsService .
func NewIpfsService(
	ipfsDao daoInterface.IPFSDao,
	userDao daoInterface.UserDao,
	projectDao daoInterface.ProjectDao,
	rdb redis.UniversalClient,
	cache cache.Cache,
	logger *zap.Logger,
) *IpfsService {
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

		// "/object/get",
		// "/dag/import",

		// TODO: some method can cache
		// version
		// object/get
		// block state

	}
	return &IpfsService{
		ipfsDao:     ipfsDao,
		userDao:     userDao,
		projectDao:  projectDao,
		rdb:         rdb,
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

func (s *IpfsService) GetUserIpfsFile(ctx context.Context, apiKey, cid string) (models.IPFSFile, error) {
	userID := s.getUserIDByAPIKey(ctx, apiKey)
	var file models.IPFSFile
	if err := s.ipfsDao.GetUserFile(userID, cid, &file); err != nil {
		return models.IPFSFile{}, err
	}
	return file, nil
}

func (s *IpfsService) PinObject(ctx context.Context, apiKey, cid string) error {
	return s.ipfsDao.SetUserFilePinStatus(s.getUserIDByAPIKey(ctx, apiKey), models.PinStatusPin, cid)
}

func (s *IpfsService) UnPinObject(ctx context.Context, apiKey, cid string) error {
	return s.ipfsDao.SetUserFilePinStatus(s.getUserIDByAPIKey(ctx, apiKey), models.PinStatusUnPin, cid)
}

// Add file info to the database
func (s *IpfsService) Add(ctx context.Context, apiKey string, addParam request.AddParam, fileList []response.AddResp) error {
	userID := s.getUserIDByAPIKey(ctx, apiKey)
	chainID := constant.ChainIPFS.ChainID
	logger := s.logger.With(zap.String("apiKey", apiKey), zap.Int("userID", userID), zap.Uint8("chainId", chainID))

	var dbFiles []*models.IPFSFile
	var dirSize uint64
	if addParam.WrapWithDirectory {
		dirSize = cast.ToUint64(fileList[len(fileList)-1].Size)
	}
	for _, v := range fileList {
		dbFile := &models.IPFSFile{
			PinStatus: addParam.PinStatus,
			UserId:    userID,
			FileSize:  cast.ToUint64(v.Size),
			FileName:  v.Name,
			CID:       v.Hash,
		}
		if addParam.WrapWithDirectory {
			dbFile.WrapDirName = fileList[len(fileList)-1].Name
			dbFile.WrapWithDirCid = fileList[len(fileList)-1].Hash
		}
		dbFiles = append(dbFiles, dbFile)
	}
	var addStorage uint64
	for _, v := range dbFiles {
		id, err := s.ipfsDao.SelectUserFileID(userID, v.CID)
		if err != nil {
			logger.With(zap.String("cid", v.CID)).Error("count user file failed", zap.Error(err))
		}
		if id == 0 {
			addStorage += v.FileSize
		}
	}
	if addStorage > 0 {
		// remove dir ipfs file size
		addStorage -= dirSize
		_, err := s.ipfsDao.IncrUserStorage(userID, addStorage)
		if err == nil {
			s.IncrIPFSUsage(
				ctx,
				apiKey,
				cachekey.IpfsLimitStorageSetKey(),
				constant.ChainIPFS.ChainID,
				int64(addStorage),
			)
		}
	}

	return s.ipfsDao.BatchFistOrCreateFiles(dbFiles)
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

func (s *IpfsService) GetApiKeyByActiveGateway(ctx context.Context, subdomain string) string {
	var apiKey string
	_ = s.cache.CacheFn(ctx,
		cachekey.ApiKeyByGateway(subdomain),
		&apiKey,
		func() error {
			gateway, err := s.ipfsDao.QueryGateway(subdomain)
			if err != nil {
				return err
			}
			var user models.IPFSUser
			err = s.ipfsDao.GetIPFSUser(gateway.UserId, &user)
			if err != nil {
				return err
			}
			if user.ActiveGatewayId != gateway.ID {
				return errors.New("gateway not active")
			}
			var p models.Project
			if err = s.projectDao.DB().Where("user_id = ?", user.ID).Order("id ASC").First(&p).Error; err != nil {
				return err
			}
			apiKey = p.APIKey
			return nil
		},
	)
	return apiKey
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
		cachekey.APIKeyMeteUserID(apiKey),
		&userID,
		func() error {
			var err error
			userID, err = s.userDao.GetIDByAPIKey(apiKey)
			return err
		},
	)
	return userID
}
