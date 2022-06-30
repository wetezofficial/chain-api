package ratelimitv1

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

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
var rateLimitScript = redis.NewScript(`
local api_key     	= KEYS[1]

local chain_id 		= ARGV[1]
local day	   		= ARGV[2]
local n	   		    = tonumber(ARGV[3])
local revert	    = ARGV[4]

local sec_quota_key = "q:s:" .. chain_id .. ":" .. api_key
local day_quota_key = "q:d:" .. chain_id .. ":" .. api_key

local sec_rate_limit_key = "s:" .. chain_id .. ":" .. api_key
local day_rate_limit_key = "d:" .. chain_id .. ":" .. api_key .. ":" .. day

local sec_quota = tonumber(redis.call("GET", sec_quota_key))
if not sec_quota then
  return -1
end

local current = tonumber(redis.call("INCRBY", sec_rate_limit_key, n))
if current == n then
    redis.call("EXPIRE",sec_rate_limit_key,1)
end

if current > sec_quota then
  return -2
end

local day_quota = tonumber(redis.call("GET", day_quota_key))
if not day_quota then
  return -1
end

current = tonumber(redis.call("INCRBY", day_rate_limit_key, n))
if current == n then
    redis.call("EXPIRE", day_rate_limit_key, 129600)
end

if current > day_quota then
  if revert == "1" then
    redis.call("DECRBY", day_rate_limit_key, current - day_quota)
  end
  return -3
end

return 1
`)
