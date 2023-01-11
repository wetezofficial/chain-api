package ratelimitv1

import (
	"context"
	"fmt"
	"time"

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
	logger    *zap.Logger
	whitelist []string
}

func NewRateLimiter(rdb redis.UniversalClient, logger *zap.Logger, whitelist []string) (*RateLimiter, error) {
	return &RateLimiter{
		rdb:       rdb,
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
