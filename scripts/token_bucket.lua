-- KEYS[1...N]: The names of the buckets
-- ARGV[1]: Limits (JSON array: [[duration, max_allowed, weight], ...])
-- ARGV[2]: Current timestamp in milliseconds
local limits = cjson.decode(ARGV[1])
local now = tonumber(ARGV[2]) / 1000
local key_prefix = "app:rate-limiter:token-bucket:"

local updates = {}

for i, limit in ipairs(limits) do
    local duration = tonumber(limit[1])
    local max_allowed = tonumber(limit[2])
    local weight = tonumber(limit[3])

    for j, key in ipairs(KEYS) do
        local ks = key_prefix .. key .. ":" .. duration

        local map = redis.call('HMGET', ks, "count", "ts")
        local current = tonumber(map[1]) or max_allowed
        local last_ts = tonumber(map[2]) or now

        local fill_rate = max_allowed / duration
        
        local tokens_to_add = (now - last_ts) * fill_rate
        local value = math.min(current + tokens_to_add, max_allowed)

        if value < weight then
            local missing = weight - value
            local wait_time = missing / fill_rate
            return {0, 0, math.ceil(wait_time)}
        end

        table.insert(updates, {
            key = ks,
            balance = value - weight,
            ttl = math.ceil(duration) 
        })
    end
end

-- Initialize remaining to a safe default if no updates were queued
local remaining = updates[1] and updates[1].balance or 0

-- If request is allowed, apply all queued updates
for _, update in ipairs(updates) do
    redis.call('HMSET', update.key, "count", update.balance, "ts", now)
    redis.call('EXPIRE', update.key, update.ttl)
    remaining = math.min(remaining, update.balance)
end

-- Return the floor of remaining so the client sees clean integers
return {1, math.floor(remaining), 0}