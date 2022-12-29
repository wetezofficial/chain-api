package ratelimitv1

import (
	"context"
	"errors"
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

	BandWidthUpload   = 1
	BandWidthDownload = 2
)

// 5gb
var limitMonthBandwidth int64 = 1024 * 1024 * 1024 * 5

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
		if err = l.rdb.IncrBy(ctx, cachekey.GetChainDayKey(chainID, t), int64(n)).Err(); err != nil {
			logger.Error("failed to save chain day quota", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetChainHourKey(chainID), int64(n)).Err(); err != nil {
			logger.Error("failed to save chain newest hour quota", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetChainTotalKey(chainID), int64(n)).Err(); err != nil {
			logger.Error("failed to save chain day quota", zap.Error(err))
		}
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

	inWhitelist, err := l.allowWhitelist(ctx, chainID, apiKey, 1)
	if err != nil {
		return err
	}
	if inWhitelist {
		return nil
	}

	t := time.Now()

	// TODO: the limit rule

	switch bwType {
	case BandWidthUpload:
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWDayUpKey(chainID, t), fileSize).Err(); err != nil {
			logger.Error("failed to save chain day bandwidth:", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWHourUpKey(chainID), fileSize).Err(); err != nil {
			logger.Error("failed to save chain month bandwidth:", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWMonthUpKey(chainID, t), fileSize).Err(); err != nil {
			logger.Error("failed to save chain newest hour quota", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWTotalUpKey(chainID), fileSize).Err(); err != nil {
			logger.Error("failed to save chain day quota", zap.Error(err))
		}
	case BandWidthDownload:
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWDayDownKey(chainID, t), fileSize).Err(); err != nil {
			logger.Error("failed to save chain day bandwidth:", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWHourDownKey(chainID), fileSize).Err(); err != nil {
			logger.Error("failed to save chain month bandwidth:", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWMonthDownKey(chainID, t), fileSize).Err(); err != nil {
			logger.Error("failed to save chain newest hour quota", zap.Error(err))
		}
		if err = l.rdb.IncrBy(ctx, cachekey.GetBWTotalDownKey(chainID), fileSize).Err(); err != nil {
			logger.Error("failed to save chain day quota", zap.Error(err))
		}
	default:
		return errors.New("unsupported type")
	}

	return nil
}
