package ratelimitv1

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

//go:embed lua/ratelimit.lua
var ratelimitLuaScript string

func RedisAllow(ctx context.Context, rdb redis.Scripter, chainID uint8, apiKey string, t time.Time, n int, revert bool) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("n must be greater than 0")
	}

	r := "0"
	if revert {
		r = "1"
	}

	// use api key as hashtag for sharding
	apiKey = "{" + apiKey + "}"
	return rateLimitScript.Run(ctx, rdb, []string{apiKey}, chainID, t.Day(), n, r).Int()
}

// rateLimitScript is a redis lua script for throttling
//   day_rate_limit_key expire in 1.5 days
//
// ATTENTION: tonumber(nil) is still (nil) not 0, @see http://lua-users.org/lists/lua-l/2009-04/msg00370.html
//
// INPUT: NO KEYS, 3 arguments
// return values
// 1 OK
// -1 key not exist
// -2 exceed second limitation
// -3 exceed day limitation
var rateLimitScript = redis.NewScript(ratelimitLuaScript)
