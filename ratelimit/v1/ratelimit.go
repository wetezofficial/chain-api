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

func (l *RateLimiter) Allow(ctx context.Context, chainID uint8, apiKey string) error {
	logger := l.logger.With(zap.String("apiKey", apiKey), zap.Uint8("chainId", chainID))

	apiKey = "{" + apiKey + "}" // use api key as hashtag for sharding
	res, err := RateLimitScript.Run(ctx, l.rdb, []string{apiKey}, chainID, time.Now().Day()).Int()
	if err != nil {
		logger.Error("failed to run rate limit script", zap.Error(err))
		return err
	}

	if l.skipper != nil && l.skipper(chainID, apiKey[1:len(apiKey)-1]) {
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
