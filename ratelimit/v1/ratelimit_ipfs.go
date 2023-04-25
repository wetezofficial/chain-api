package ratelimitv1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"starnet/chain-api/pkg/utils"
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
		utils.IncreaseAndSetExpire(ctx, l.rdb, cachekey.GetUserBWDayUpKey(apiKey, chainID, t), fileSize, time.Hour*36, logger)
		// update transfer up user usage
		l.ipfsSrv.IncrIPFSUsage(ctx, apiKey, cachekey.IpfsLimitTransferUpSetKey(), chainID, fileSize)
	case BandWidthDownload:
		utils.IncreaseAndSetExpire(ctx, l.rdb, cachekey.GetUserBWDayDownKey(apiKey, chainID, t), fileSize, time.Hour*36, logger)
		// update transfer down user usage
		l.ipfsSrv.IncrIPFSUsage(ctx, apiKey, cachekey.IpfsLimitTransferDownSetKey(), chainID, fileSize)
	default:
		return errors.New("unsupported type")
	}

	return nil
}

func (l *RateLimiter) CheckIPFSLimit(
	ctx context.Context,
	apiKey string,
	chainID uint8,
	logger *zap.Logger,
	fileSize int64,
	bwType uint8,
) (bool, error) {
	usageRecord, err, done := l.GetIPFSUserUsage(ctx, apiKey, chainID, logger)
	if !done {
		return done, err
	}

	userLimit, err := l.rdb.HGetAll(ctx, cachekey.GetUserIpfsLimitKey(apiKey, chainID)).Result()
	if err != nil {
		errMsg := "can not read the plan limit"
		logger.Error(errMsg, zap.Error(err))
		return false, fmt.Errorf(errMsg)
	}

	for k, v := range usageRecord {
		usage := v
		if (k == cachekey.IpfsLimitStorageSetKey() || k == cachekey.IpfsLimitTransferUpSetKey()) &&
			bwType == BandWidthUpload {
			usage += uint64(fileSize)
		}
		if k == cachekey.IpfsLimitTransferDownSetKey() && bwType == BandWidthDownload {
			usage += uint64(fileSize)
		}
		if usage >= cast.ToUint64(userLimit[k]) {
			logger.Error("the %s out the plan limit")
			return false, nil
		}
	}

	return true, nil
}

// GetIPFSUserUsage query ipfs usage return map, if not exist init data
func (l *RateLimiter) GetIPFSUserUsage(ctx context.Context, apiKey string, chainID uint8, logger *zap.Logger) (map[string]uint64, error, bool) {
	result := make(map[string]uint64, 3)
	usageRecord, err := l.rdb.HGetAll(ctx, cachekey.GetUserIPFSUsageKey(apiKey, chainID)).Result()
	if err != nil {
		if models.IsNotFound(err) {
			userInfo, err := l.ipfsSrv.GetIpfsUserNoCache(ctx, apiKey)
			if err != nil {
				errMsg := "get ipfs User form db failed"
				e := fmt.Errorf(errMsg)
				logger.Error(err.Error(), zap.Error(err))
				return nil, e, true
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
		} else {
			errorMsg := "can not read the user ipfs auth status"
			logger.Error(errorMsg, zap.Error(err))
			return nil, err, false
		}
	}
	for k, v := range usageRecord {
		result[k] = cast.ToUint64(v)
	}
	return result, nil, true
}
