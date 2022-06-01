package ratelimitv1

//func TestRateLimitDay(t *testing.T) {
//	rdb := redis.NewClient(&redis.Options{
//		Addr: "localhost:6379",
//	})
//	ctx := context.TODO()
//	rdb.FlushDB(ctx)
//
//	apikey := "test_rate_limit_day_apikey"
//	// Initialize configuration
//	numPerSecond := 1000
//	rdb.Set(ctx, fmt.Sprintf("conf:%s:s", apikey), numPerSecond, 0)
//	rdb.Set(ctx, fmt.Sprintf("conf:%s:d", apikey), numPerSecond*5, 0)
//
//	for count := 1; count <= 5; count++ {
//		for i := 0; i < numPerSecond; i++ {
//			res, err := RateLimitScript.Run(ctx, rdb, []string{apikey}).Int()
//			assert.Nil(t, err)
//			assert.Equal(t, Allow, res)
//			time.Sleep(time.Second / time.Duration(numPerSecond))
//		}
//
//		dayLimitKey := "day:" + apikey
//		dayLimit, err := rdb.Get(ctx, dayLimitKey).Int()
//		assert.Nil(t, err)
//		assert.Equal(t, numPerSecond*count, dayLimit)
//	}
//
//	for i := 0; i < numPerSecond*5; i++ {
//		res, err := RateLimitScript.Run(ctx, rdb, nil, apikey).Int()
//		assert.Nil(t, err)
//		assert.GreaterOrEqual(t, ExceedSecondLimit, res)
//	}
//}
//
//func TestRateLimitSecond(t *testing.T) {
//	rdb := redis.NewClient(&redis.Options{
//		Addr: "localhost:6379",
//	})
//	ctx := context.TODO()
//	rdb.FlushDB(ctx)
//
//	apikey := "key1"
//	// Initialize configuration
//	rdb.Set(ctx, fmt.Sprintf("conf:%s:s", apikey), "5", 0)
//	rdb.Set(ctx, fmt.Sprintf("conf:%s:d", apikey), "5000", 0)
//
//	res, err := RateLimitScript.Run(context.TODO(), rdb, nil, "notexist").Int()
//	assert.Nil(t, err)
//	assert.Equal(t, NotExist, res)
//
//	for i := 0; i < 10; i++ {
//		res, err = RateLimitScript.Run(context.TODO(), rdb, nil, apikey).Int()
//		assert.Nil(t, err)
//		if i < 5 {
//			assert.Equal(t, Allow, res)
//		} else {
//			assert.Equal(t, ExceedSecondLimit, res)
//		}
//	}
//
//	dayLimitKey := "day:" + apikey
//	dayLimit, err := rdb.Get(ctx, dayLimitKey).Int()
//	assert.Nil(t, err)
//	assert.Equal(t, 5, dayLimit)
//
//	time.Sleep(time.Second)
//
//	for i := 0; i < 10; i++ {
//		res, err := RateLimitScript.Run(context.TODO(), rdb, nil, apikey).Int()
//		assert.Nil(t, err)
//		if i < 5 {
//			assert.Equal(t, Allow, res)
//		} else {
//			assert.Equal(t, ExceedSecondLimit, res)
//		}
//	}
//
//	dayLimit, err = rdb.Get(ctx, dayLimitKey).Int()
//	assert.Nil(t, err)
//	assert.Equal(t, 10, dayLimit)
//}
//
//func BenchmarkRateLimit(b *testing.B) {
//	rdb := redis.NewClient(&redis.Options{
//		Addr: "localhost:6379",
//	})
//	ctx := context.TODO()
//	rdb.FlushDB(ctx)
//
//	apikey := "key1"
//	// Initialize configuration
//	rdb.Set(ctx, fmt.Sprintf("conf:%s:s", apikey), "5", 0)
//	rdb.Set(ctx, fmt.Sprintf("conf:%s:d", apikey), "5000", 0)
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		res, err := RateLimitScript.Run(context.TODO(), rdb, nil, apikey).Int()
//		assert.Nil(b, err)
//		if res == Allow {
//			_ = res
//		}
//		if res == NotExist {
//			_ = res
//		}
//		if res == ExceedSecondLimit {
//			_ = res
//		}
//		if res == ExceedDayLimit {
//			_ = res
//		}
//	}
//}
