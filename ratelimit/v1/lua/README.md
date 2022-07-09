Redis Lua Scripts Test and Generate Code Coverage 
===============

### Start redis

~~~bash
$ docker compose up redis -d
~~~

### Run test code

~~~bash
$ docker compose up lua
~~~

    rm -rf luacov.stats.out ; busted -c test-ratelimit.lua
    ++++++++++
    10 successes / 0 failures / 0 errors / 0 pending : 0.193787 seconds
    rm -rf luacov.report.out ; luacov '^ratelimit' && cat luacov.report.out
    ==============================================================================
    ratelimit.lua
    ==============================================================================
     15 local api_key     	= KEYS[1]
     15 local chain_id 		= ARGV[1]
     15 local day	   		= ARGV[2]
     15 local n	   		    = tonumber(ARGV[3])
     15 local revert	    = ARGV[4]
     15 local sec_quota_key = "q:s:" .. chain_id .. ":" .. api_key
     15 local day_quota_key = "q:d:" .. chain_id .. ":" .. api_key
     15 local sec_rate_limit_key = "s:" .. chain_id .. ":" .. api_key
     15 local day_rate_limit_key = "d:" .. chain_id .. ":" .. api_key .. ":" .. day
     15 local sec_quota = tonumber(redis.call("GET", sec_quota_key))
     15 if not sec_quota then
      1   return -1
        end
     14 local current = redis.call("INCRBY", sec_rate_limit_key, n)
     14 if current == n then
      9     redis.call("EXPIRE",sec_rate_limit_key,1)
        end
     14 if current > sec_quota then
      1   return -2
        end
     13 local day_quota = tonumber(redis.call("GET", day_quota_key))
     13 if not day_quota then
      1   return -1
        end
     12 current = redis.call("INCRBY", day_rate_limit_key, n)
     12 if current == n then
      8     redis.call("EXPIRE", day_rate_limit_key, 129600)
        end
     12 if current > day_quota then
      3   if revert == "1" then
      2     redis.call("DECRBY", day_rate_limit_key, current - day_quota)
          end
      3   return -3
        end
      9 return 1
    
    ==============================================================================
    Summary
    ==============================================================================
    
    File          Hits Missed Coverage
    ----------------------------------
    ratelimit.lua 28   0      100.00%
    ----------------------------------
    Total         28   0      100.00%