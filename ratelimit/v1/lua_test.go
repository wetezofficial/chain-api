package ratelimitv1

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"starnet/starnet/dao"
	"testing"
	"time"
)

func TestCount(t *testing.T) {
	t.Parallel()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.TODO()

	var chainID uint8 = 1
	rateLimitDao := dao.NewRateLimitDao(rdb)

	genAndSetupApikey := func(secQuota, dayQuota int) string {
	FnBegin:
		apikey, err := uuid.GenerateUUID()
		assert.Nil(t, err)

		// Initialize configuration
		err = rateLimitDao.SetQuota(apikey, int(chainID), secQuota, dayQuota)
		assert.Nil(t, err)

		_, err = rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		if !errors.Is(err, redis.Nil) {
			// 若key已经存在，则重新生成一个作为测试
			goto FnBegin
		}
		return apikey
	}

	t.Run("test with count", func(t *testing.T) {
		secQuota := 10
		dayQuota := 50
		apikey := genAndSetupApikey(secQuota, dayQuota)

		total := 0
		for i := 0; i < 5; i++ {
			n := i%secQuota + 1
			total += n
			res, err := RedisAllow(ctx, rdb, chainID, apikey, time.Now(), n, true)
			assert.Nil(t, err)
			_ = res
			time.Sleep(time.Second / time.Duration(secQuota) * time.Duration(n))
		}

		dayUsage, err := rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		assert.Nil(t, err)
		assert.Equal(t, int64(total), dayUsage)
	})
}

func TestRevert(t *testing.T) {
	t.Parallel()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.TODO()

	var chainID uint8 = 1
	rateLimitDao := dao.NewRateLimitDao(rdb)

	genAndSetupApikey := func(secQuota, dayQuota int) string {
	FnBegin:
		apikey, err := uuid.GenerateUUID()
		assert.Nil(t, err)

		// Initialize configuration
		err = rateLimitDao.SetQuota(apikey, int(chainID), secQuota, dayQuota)
		assert.Nil(t, err)

		_, err = rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		if !errors.Is(err, redis.Nil) {
			// 若key已经存在，则重新生成一个作为测试
			goto FnBegin
		}
		return apikey
	}

	t.Run("test revert", func(t *testing.T) {
		secQuota := 50
		dayQuota := 50
		exceed := 5
		apikey := genAndSetupApikey(secQuota, dayQuota)

		for i := 0; i < dayQuota+exceed; i++ {
			res, err := RedisAllow(ctx, rdb, chainID, apikey, time.Now(), 1, true)
			assert.Nil(t, err)
			_ = res
			time.Sleep(time.Second / time.Duration(secQuota))
		}

		dayUsage, err := rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		assert.Nil(t, err)
		assert.Equal(t, int64(dayQuota), dayUsage)
	})

	t.Run("test not revert", func(t *testing.T) {
		secQuota := 50
		dayQuota := 50
		apikey := genAndSetupApikey(secQuota, dayQuota)

		var i = 0
		for ; i < dayQuota+5; i++ {
			res, err := RedisAllow(ctx, rdb, chainID, apikey, time.Now(), 1, false)
			assert.Nil(t, err)
			_ = res
			time.Sleep(time.Second / time.Duration(secQuota))
		}

		dayUsage, err := rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		assert.Nil(t, err)
		assert.Equal(t, int64(i), dayUsage)
	})
}

func TestExceeded(t *testing.T) {
	t.Parallel()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.TODO()

	var chainID uint8 = 1
	rateLimitDao := dao.NewRateLimitDao(rdb)

	genAndSetupApikey := func(secQuota, dayQuota int) string {
	FnBegin:
		apikey, err := uuid.GenerateUUID()
		assert.Nil(t, err)

		// Initialize configuration
		err = rateLimitDao.SetQuota(apikey, int(chainID), secQuota, dayQuota)
		assert.Nil(t, err)

		_, err = rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		if !errors.Is(err, redis.Nil) {
			// 若key已经存在，则重新生成一个作为测试
			goto FnBegin
		}
		return apikey
	}

	t.Run("test exceeded second quota", func(t *testing.T) {
		secQuota := 10
		apikey := genAndSetupApikey(secQuota, 1000)

		now := time.Now()
		for i := 0; i < secQuota; i++ {
			res, err := RedisAllow(ctx, rdb, chainID, apikey, now, 1, true)
			assert.Nil(t, err)
			assert.Equal(t, Allow, res)
		}

		res, err := RedisAllow(ctx, rdb, chainID, apikey, now, 1, true)
		assert.Nil(t, err)
		assert.Equal(t, ExceedSecondLimit, res)
	})

	t.Run("test exceeded day quota", func(t *testing.T) {
		secQuota := 200
		dayQuota := 500
		apikey := genAndSetupApikey(secQuota, dayQuota)

		for i := 0; i < dayQuota; i++ {
			res, err := RedisAllow(ctx, rdb, chainID, apikey, time.Now(), 1, true)
			assert.Nil(t, err)
			assert.Equal(t, Allow, res)
			time.Sleep(time.Second / time.Duration(secQuota))
		}

		res, err := RedisAllow(ctx, rdb, chainID, apikey, time.Now(), 1, true)
		assert.Nil(t, err)
		assert.Equal(t, ExceedDayLimit, res)
	})

	t.Run("test get day usage", func(t *testing.T) {
		secQuota := 200
		dayQuota := 500
		apikey := genAndSetupApikey(secQuota, dayQuota)
		n := rand.Intn(dayQuota - 1)

		for i := 0; i < n; i++ {
			res, err := RedisAllow(ctx, rdb, chainID, apikey, time.Now(), 1, true)
			assert.Nil(t, err)
			assert.Equal(t, Allow, res)
			time.Sleep(time.Second / time.Duration(secQuota))
		}

		dayUsage, err := rateLimitDao.GetDayUsage(apikey, int(chainID), time.Now())
		assert.Nil(t, err)
		assert.Equal(t, int64(n), dayUsage)
	})
}
