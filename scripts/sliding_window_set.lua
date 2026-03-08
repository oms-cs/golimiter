-- KEYS: Bucket names
-- ARGV[1]: Limits JSON [[duration, max_allowed, weight], ...]
-- ARGV[2]: Current timestamp in milliseconds
local limits = cjson.decode(ARGV[1])
local now_ms = tonumber(ARGV[2])
local now_sec = now_ms / 1000

local key_prefix = "app:rate-limiter:sliding-zset:"
local check_results = {}
local max_wait_time = 0

for i, limit in ipairs(limits) do
    local duration = tonumber(limit[1])
    local max_allowed = tonumber(limit[2])
    local weight = tonumber(limit[3] or 1)
    
    for _, base_key in ipairs(KEYS) do
        local ks = key_prefix .. base_key .. ":" .. duration
        local window_start = now_sec - duration

        redis.call('ZREMRANGEBYSCORE', ks, '-inf', window_start)

        local current_count = redis.call('ZCARD', ks)

        if current_count + weight > max_allowed then
            local oldest = redis.call('ZRANGE', ks, 0, 0, 'WITHSCORES')
            local tier_wait = 0
            if oldest[2] then
                tier_wait = tonumber(oldest[2]) + duration - now_sec
            else
                tier_wait = duration
            end
            max_wait_time = math.max(max_wait_time, tier_wait)
        end
        
        table.insert(check_results, {
            key = ks,
            duration = duration,
            weight = weight,
            remaining = math.max(0, max_allowed - (current_count + weight))
        })
    end
end

if max_wait_time > 0 then
    return {0, 0, math.ceil(max_wait_time)}
end

local global_min_remaining = 999999999

for i, res in ipairs(check_results) do
    for w = 1, res.weight do
        local member_id = now_ms .. ":" .. i .. ":" .. w
        redis.call('ZADD', res.key, now_sec, member_id)
    end
    
    redis.call('EXPIRE', res.key, math.ceil(res.duration))
    
    if res.remaining < global_min_remaining then
        global_min_remaining = res.remaining
    end
end

return {1, global_min_remaining, 0}