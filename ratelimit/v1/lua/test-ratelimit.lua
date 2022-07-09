local call_redis_script = require "./harness";

local chain_id = "1"
local api_key = "{{test_api_key}}"
local day = "01"

local sec_quota_key = "q:s:" .. chain_id .. ":" .. api_key
local day_quota_key = "q:d:" .. chain_id .. ":" .. api_key

local sec_rate_limit_key = "s:" .. chain_id .. ":" .. api_key
local day_rate_limit_key = "d:" .. chain_id .. ":" .. api_key .. ":" .. day

function ratelimit(api_key, chain_id, day, n, revert)
  return call_redis_script('ratelimit.lua',  { api_key }, { chain_id, day, n, revert });
end

describe("ratelimit", function()

  -- Flush the database before running the tests
  before_each(function()
    redis.call('FLUSHDB')
  end)

  it("should not allow when second quota is not set", function()
    local result = ratelimit(api_key, chain_id, "01", "1", "1")
    assert.are.equals(-1, result)
  end)

  it("should not allow when day quota is not set", function()
    local result = redis.call("SET", sec_quota_key, "1")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, "01", "1", "1")
    assert.are.equals(-1, result)
  end)

  it("should have expiration of second rate limit key", function()
    local result = redis.call("SET", sec_quota_key, "1")
    assert.are.same(true, result)
    local result = redis.call("SET", day_quota_key, "100")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "1", "1")
    assert.are.same(1, result)

    local result = redis.call("TTL", sec_rate_limit_key)
    assert.are.same(1, result)
  end)

  it("should have expiration of second rate limit key with n", function()
    local result = redis.call("SET", sec_quota_key, "2")
    assert.are.same(true, result)
    local result = redis.call("SET", day_quota_key, "100")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "2", "1")
    assert.are.same(1, result)

    local result = redis.call("TTL", sec_rate_limit_key)
    assert.are.same(1, result)
  end)

  it("should have expiration of day rate limit key", function()
    local result = redis.call("SET", sec_quota_key, "1")
    assert.are.same(true, result)
    local result = redis.call("SET", day_quota_key, "100")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "1", "1")
    assert.are.same(1, result)

    local result = redis.call("TTL", day_rate_limit_key)
    assert.are.same(129600, result)
  end)

  it("should have expiration of day rate limit key with n", function()
    local result = redis.call("SET", sec_quota_key, "2")
    assert.are.same(true, result)
    local result = redis.call("SET", day_quota_key, "100")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "2", "1")
    assert.are.same(1, result)

    local result = redis.call("TTL", day_rate_limit_key)
    assert.are.same(129600, result)
  end)

  it("should exceed second rate quota", function()
    local result = redis.call("SET", sec_quota_key, "1")
    assert.are.same(true, result)

    local result = redis.call("SET", day_quota_key, "100")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "1", "1")
    assert.are.same(1, result)

    local result = ratelimit(api_key, chain_id, day, "1", "1")
    assert.are.same(-2, result)
  end)

  it("should revert day rate limit", function()
    local result = redis.call("SET", sec_quota_key, "2")
    assert.are.same(true, result)

    local result = redis.call("SET", day_quota_key, "1")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "1", "1")
    assert.are.same(1, result)

    local result = ratelimit(api_key, chain_id, day, "1", "1")
    assert.are.same(-3, result)

    local result = tonumber(redis.call("GET", day_rate_limit_key))
    assert.are.same(1, result)
  end)

  it("should not revert day rate limit", function()
    local result = redis.call("SET", sec_quota_key, "2")
    assert.are.same(true, result)

    local result = redis.call("SET", day_quota_key, "1")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "1", "0")
    assert.are.same(1, result)

    local result = ratelimit(api_key, chain_id, day, "1", "0")
    assert.are.same(-3, result)

    local result = tonumber(redis.call("GET", day_rate_limit_key))
    assert.are.same(2, result)
  end)

  it("should work properly with custom n", function()
    local result = redis.call("SET", sec_quota_key, "6")
    assert.are.same(true, result)

    local result = redis.call("SET", day_quota_key, "4")
    assert.are.same(true, result)

    local result = ratelimit(api_key, chain_id, day, "2", "1")
    assert.are.same(1, result)

    local result = ratelimit(api_key, chain_id, day, "2", "1")
    assert.are.same(1, result)

    local result = ratelimit(api_key, chain_id, day, "2", "1")
    assert.are.same(-3, result)

    local result = tonumber(redis.call("GET", day_rate_limit_key))
    assert.are.same(4, result)
  end)
end)
