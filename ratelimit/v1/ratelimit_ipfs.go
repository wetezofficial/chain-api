package ratelimitv1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"starnet/starnet/cachekey"
	"starnet/starnet/models"

	"github.com/spf13/cast"

	"go.uber.org/zap"
)

const (
	BandWidthUpload   = 1
	BandWidthDownload = 2
)

// BandwidthHook update the chain (ipfs) bandwidth periodic usage
func (l *RateLimiter) BandwidthHook(
	ctx context.Context,
	chainID uint8,
	apiKey string,
	fileSize int64,
	bwType uint8,
) error {
	logger := l.logger.With(zap.String("apiKey", apiKey), zap.Uint8("chainId", chainID))

	if inWhitelist, err := l.allowWhitelist(ctx, chainID, apiKey, 1); err != nil {
		return err
	} else if inWhitelist {
		return nil
	}

	t := time.Now()

	switch bwType {
	case BandWidthUpload:
		_ = l.increaseUserUpBandwidth(ctx, chainID, apiKey, t, fileSize, logger)
	case BandWidthDownload:
		_ = l.increaseUserDownBandwidth(ctx, chainID, apiKey, t, fileSize, logger)
	default:
		return errors.New("unsupported type")
	}

	return nil
}

// CheckIPFSLimit TODO: 思考 CheckIPFSLimit 可不可以和 bandHook 中的 incr 合并到一起
func (l *RateLimiter) CheckIPFSLimit(
	ctx context.Context,
	apiKey string,
	chainID uint8,
	logger *zap.Logger,
	fileSize int64,
	bwType uint8,
) error {
	usageRecord, err := l.rdb.HGetAll(ctx, cachekey.GetUserIPFSUsageKey(apiKey, chainID)).Result()
	if err != nil {
		if models.IsNotFound(err) {
			userInfo, err := l.ipfsSrv.GetIpfsUserNoCache(ctx, apiKey)
			if err != nil {
				errMsg := "get ipfs User form db failed"
				e := fmt.Errorf(errMsg)
				logger.Error(errMsg, zap.Error(e))
				return e
			}

			if err = l.rdb.HSet(
				context.TODO(),
				cachekey.GetUserIPFSUsageKey(apiKey, chainID),
				cachekey.IpfsLimitStorageSetKey(),
				userInfo.TotalStorage,
			).Err(); err != nil {
				logger.Error("set user auth failed", zap.Error(err))
			}
			if err = l.rdb.HSet(
				context.TODO(),
				cachekey.GetUserIPFSUsageKey(apiKey, chainID),
				cachekey.IpfsLimitTransferUpSetKey(),
				userInfo.TransferUp,
			).Err(); err != nil {
				logger.Error("set user auth failed", zap.Error(err))
			}
			if err = l.rdb.HSet(
				context.TODO(),
				cachekey.GetUserIPFSUsageKey(apiKey, chainID),
				cachekey.IpfsLimitTransferDownSetKey(),
				userInfo.TransferDown,
			).Err(); err != nil {
				logger.Error("set user auth failed", zap.Error(err))
			}
			return nil
		}
		errorMsg := "can not read the user ipfs auth status"
		logger.Error(errorMsg, zap.Error(err))
		return fmt.Errorf(errorMsg)
	}

	userLimit, err := l.rdb.HGetAll(ctx, cachekey.GetUserIpfsLimitKey(apiKey, chainID)).Result()
	if err != nil {
		errMsg := "can not read the plan limit"
		logger.Error(errMsg, zap.Error(err))
		return fmt.Errorf(errMsg)
	}

	for k, v := range usageRecord {
		usage := cast.ToUint64(v)
		if (k == cachekey.IpfsLimitStorageSetKey() || k == cachekey.IpfsLimitTransferUpSetKey()) &&
			bwType == BandWidthUpload {
			usage += uint64(fileSize)
		}
		if k == cachekey.IpfsLimitTransferDownSetKey() && bwType == BandWidthDownload {
			usage += uint64(fileSize)
		}
		if cast.ToUint64(v) >= cast.ToUint64(userLimit[k]) {
			return fmt.Errorf("The %s out the plan limit", k)
		}
	}

	return nil
}

func (l *RateLimiter) increaseUserUpBandwidth(ctx context.Context, chainID uint8, apiKey string, t time.Time, fileSize int64, logger *zap.Logger) error {
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWHourUpKey(apiKey, chainID), fileSize, time.Minute*90, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWDayUpKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWMonthUpKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)

	// update transfer up user usage
	l.ipfsSrv.IncrIPFSUsage(ctx, apiKey, cachekey.IpfsLimitTransferUpSetKey(), chainID, fileSize)

	return nil
}

func (l *RateLimiter) increaseUserDownBandwidth(ctx context.Context, chainID uint8, apiKey string, t time.Time, fileSize int64, logger *zap.Logger) error {
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWHourDownKey(apiKey, chainID), fileSize, time.Minute*90, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWDayDownKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)
	l.increaseAndSetExpire(ctx, cachekey.GetUserBWMonthDownKey(apiKey, chainID, t), fileSize, time.Second*129600, logger)

	// update transfer down user usage
	l.ipfsSrv.IncrIPFSUsage(ctx, apiKey, cachekey.IpfsLimitTransferDownSetKey(), chainID, fileSize)

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
