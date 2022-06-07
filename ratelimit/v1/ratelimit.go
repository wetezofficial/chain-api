package ratelimitv1

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

const (
	Allow             int = 1
	NotExist          int = -1
	ExceedSecondLimit int = -2
	ExceedDayLimit    int = -3
)

type RateLimiter struct {
	rdb     redis.UniversalClient
	logger  *zap.Logger
	skipper func(chainID uint8, apiKey string) bool
}

func NewRateLimiter(rdb redis.UniversalClient, logger *zap.Logger, skipper func(chainID uint8, apiKey string) bool) (*RateLimiter, error) {
	return &RateLimiter{
		rdb:     rdb,
		logger:  logger,
		skipper: skipper,
	}, nil
}

var ExceededRateLimitError = fmt.Errorf("exceeded rate limit")
var ApiKeyNotExistError = fmt.Errorf("api key not exist")

func (l *RateLimiter) Allow(ctx context.Context, chainID uint8, apiKey string, n int) error {
	logger := l.logger.With(zap.String("apiKey", apiKey), zap.Uint8("chainId", chainID))

	revert := true // 若使用超限，则 revert 超过的部分，不然显示的时候会出现使用量大于限制的情况
	whitelist := false
	if l.skipper != nil && l.skipper(chainID, apiKey) {
		whitelist = true
		revert = false
	}

	res, err := RedisAllow(ctx, l.rdb, chainID, apiKey, time.Now(), n, revert)
	if err != nil {
		logger.Error("failed to run rate limit script", zap.Error(err))
		return err
	}

	if whitelist {
		return nil
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
