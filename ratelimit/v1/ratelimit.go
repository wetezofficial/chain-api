package ratelimitv1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"starnet/chain-api/service"
	"starnet/starnet/cachekey"
	"starnet/starnet/models"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"

	"go.uber.org/zap"
)

const (
	Allow             int = 1
	NotExist          int = -1
	ExceedSecondLimit int = -2
	ExceedDayLimit    int = -3

	BandWidthUpload   = 1
	BandWidthDownload = 2
)

// 5gb
var limitMonthBandwidth int64 = 1024 * 1024 * 1024 * 5

type RateLimiter struct {
	rdb       redis.UniversalClient
	ipfsSrv   *service.IpfsService
	logger    *zap.Logger
	whitelist []string
}

func NewRateLimiter(rdb redis.UniversalClient, ipfsSrv *service.IpfsService, logger *zap.Logger, whitelist []string) (*RateLimiter, error) {
	return &RateLimiter{
		rdb:       rdb,
		ipfsSrv:   ipfsSrv,
		logger:    logger,
		whitelist: whitelist,
	}, nil
}

var (
	ExceededRateLimitError = fmt.Errorf("exceeded rate limit")
	ApiKeyNotExistError    = fmt.Errorf("api key not exist")
)

func (l *RateLimiter) allowWhitelist(ctx context.Context, chainID uint8, apiKey string, n int) (bool, error) {
	inWhitelist := false
	for _, _apiKey := range l.whitelist {
		if _apiKey == apiKey {
			inWhitelist = true
			break
		}
	}

	if inWhitelist {
		key := fmt.Sprintf("d:%d:{%s}:%d", chainID, apiKey, time.Now().Day())
		count, err := l.rdb.IncrBy(ctx, key, int64(n)).Result()
		if err != nil {
			return inWhitelist, err
		}
		if count == int64(n) {
			err = l.rdb.Expire(ctx, key, time.Hour*36).Err() // 1.5 days
			if err != nil {
				return inWhitelist, err
			}
		}
	}

	return inWhitelist, nil
}

func (l *RateLimiter) CheckInWhiteList(apiKey string) bool {
	inWhitelist := false
	for _, _apiKey := range l.whitelist {
		if _apiKey == apiKey {
			inWhitelist = true
			break
		}
	}
	return inWhitelist
}

func (l *RateLimiter) Allow(ctx context.Context, chainID uint8, apiKey string, n int) error {
	logger := l.logger.With(zap.String("apiKey", apiKey), zap.Uint8("chainId", chainID))

	inWhitelist, err := l.allowWhitelist(ctx, chainID, apiKey, n)
	if err != nil {
		return err
	}
	if inWhitelist {
		return nil
	}

	t := time.Now()

	res, err := RedisAllow(ctx, l.rdb, chainID, apiKey, t, n, true)
	if err != nil {
		logger.Error("failed to run rate limit script", zap.Error(err))
		return err
	}

	// add chain request count
	if res == 1 && n > 0 {
		l.increaseAndSetExpire(ctx, cachekey.GetChainHourKey(chainID, t), int64(n), time.Minute*90, logger)
		l.increaseAndSetExpire(ctx, cachekey.GetChainDayKey(chainID, t), int64(n), time.Hour*36, logger)

		l.increaseAndSetExpire(ctx, cachekey.GetTotalQuotaHourKey(t), int64(n), time.Minute*90, logger)
		l.increaseAndSetExpire(ctx, cachekey.GetTotalQuotaDayKey(t), int64(n), time.Hour*36, logger)
	}

	if res == NotExist {
		logger.Debug("api key not exist")
		return ApiKeyNotExistError
	}

	if res != Allow {
		logger.Debug("exceeded rate limit", zap.Int("result", res))
		return ExceededRateLimitError
	}

	return nil
}

func (l *RateLimiter) BandwidthHook(ctx context.Context, chainID uint8, apiKey string, fileSize int64, bwType uint8) error {
	logger := l.logger.With(zap.String("apiKey", apiKey), zap.Uint8("chainId", chainID))

	if inWhitelist, err := l.allowWhitelist(ctx, chainID, apiKey, 1); err != nil {
		return err
	} else if inWhitelist {
		return nil
	}

	t := time.Now()

	switch bwType {
	case BandWidthUpload:
		// TODO: check upload & totalSaveLimit
		_ = l.increaseUserUpBandwidth(ctx, chainID, apiKey, t, fileSize, logger)
	case BandWidthDownload:
		// TODO: check download & totalSaveLimit
		_ = l.increaseUserDownBandwidth(ctx, chainID, apiKey, t, fileSize, logger)
	default:
		return errors.New("unsupported type")
	}

	return nil
}

func (l *RateLimiter) CheckIpfsStorageAndUpAuth() bool {
	l.ipfsSrv.ListUserFile(ctx context.Context, apiKey string, files *[]models.IPFSFile)(ctx context.Context, apiKey string, apiParam request.AddParam, multiFileR *files.MultiFileReader)
}

func (l *RateLimiter) CheckIPFSLimit(ctx context.Context, apiKey string, chainID uint8, logger *zap.Logger) error {
	var err error

	authRecord, err := l.rdb.HGetAll(ctx, cachekey.GetUserAccessAuth(apiKey, chainID)).Result()
	if err != nil {
		if models.IsNotFound(err) {
			if err = l.rdb.HSet(context.TODO(), cachekey.GetUserAccessAuth(apiKey, chainID), cachekey.IpfsLimitStorageSetKey(), true).Err(); err != nil {
				logger.Error("set user auth failed", zap.Error(err))
			}
			if err = l.rdb.HSet(context.TODO(), cachekey.GetUserAccessAuth(apiKey, chainID), cachekey.IpfsLimitTransferUpSetKey(), true).Err(); err != nil {
				logger.Error("set user auth failed", zap.Error(err))
			}
			if err = l.rdb.HSet(context.TODO(), cachekey.GetUserAccessAuth(apiKey, chainID), cachekey.IpfsLimitTransferDownSetKey(), true).Err(); err != nil {
				logger.Error("set user auth failed", zap.Error(err))
			}
			return nil
		}
		errorMsg := "can not read the user ipfs auth status"
		logger.Error(errorMsg, zap.Error(err))
		return fmt.Errorf(errorMsg)
	}
	for k, v := range authRecord {
		if !cast.ToBool(v) {
			return fmt.Errorf("out of %s limit", k)
		}
	}

	return nil
}

func (l *RateLimiter) UpdateIPFSAuth(ctx context.Context, apiKey, setKey string, chainID uint8, status bool, logger *zap.Logger) {
	if err := l.rdb.HSet(context.TODO(), cachekey.GetUserAccessAuth(apiKey, chainID), setKey, status).Err(); err != nil {
		logger.Error("set user auth failed", zap.Error(err))
	}
}

func (l *RateLimiter) increaseUserUpBandwidth(ctx context.Context, chainID uint8, apiKey string, t time.Time, fileSize int64, logger *zap.Logger) error {
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWHourUpKey(apiKey, chainID), fileSize, time.Minute*90, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWDayUpKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWMonthUpKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)

	// TODO:rule  refactor total get from ipfsUserInfo
	return nil
}

func (l *RateLimiter) increaseUserDownBandwidth(ctx context.Context, chainID uint8, apiKey string, t time.Time, fileSize int64, logger *zap.Logger) error {
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWHourDownKey(apiKey, chainID), fileSize, time.Minute*90, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWDayDownKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWMonthDownKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)

	return nil
}

func (l *RateLimiter) increaseAndSetExpire(ctx context.Context, key string, value int64, expireTime time.Duration, logger *zap.Logger) {
	result, err := l.rdb.IncrBy(ctx, key, value).Result()
	if err != nil {
		logger.Error("failed to save key:", zap.Any(key, err))
		return
	}
	if result == value && expireTime > 0 {
		if err != l.rdb.Expire(ctx, key, expireTime).Err() {
			logger.Error("failed to save key expire:", zap.Any(key, err))
		}
	}
}
