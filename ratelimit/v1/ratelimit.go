package ratelimitv1

import (
	"context"
	"fmt"
	"time"

	"starnet/chain-api/pkg/utils"
	"starnet/chain-api/service"
	"starnet/starnet/cachekey"

	"github.com/go-redis/redis/v8"

	"go.uber.org/zap"
)

const (
	Allow             int = 1
	NotExist          int = -1
	ExceedSecondLimit int = -2
	ExceedDayLimit    int = -3
)

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

func (l *RateLimiter) ErigonCount(ctx context.Context, n int) {
	if err := l.rdb.IncrBy(ctx, cachekey.GetErigonEthTotalKey(), int64(n)).Err(); err != nil {
		l.logger.Error("ErigonCount IncrBy Error:", zap.Error(err))
	}
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
		utils.IncreaseAndSetExpire(ctx, l.rdb, cachekey.GetChainHourKey(chainID, t), int64(n), time.Minute*90, logger)
		utils.IncreaseAndSetExpire(ctx, l.rdb, cachekey.GetChainDayKey(chainID, t), int64(n), time.Hour*36, logger)

		utils.IncreaseAndSetExpire(ctx, l.rdb, cachekey.GetTotalQuotaHourKey(t), int64(n), time.Minute*90, logger)
		utils.IncreaseAndSetExpire(ctx, l.rdb, cachekey.GetTotalQuotaDayKey(t), int64(n), time.Hour*36, logger)
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
