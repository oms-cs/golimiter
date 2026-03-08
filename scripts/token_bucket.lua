-- KEYS[1]: The name of the bucket (e.g., "ratelimit:user123")
-- ARGV[1]: Limits (requests allowed per second rules json)
-- ARGV[2]: Current timestamp (Unix seconds or milliseconds)
local limits = cjson.decode(ARGV[1])

local now = tonumber(ARGV[2]) / 1000

local key_prefix = "app:rate-limiter:token-bucket:"

local updates = {}

for i, limit in ipairs(limits) do
    local duration = limit[1]
    local max_allowed = limit[2]
    local weight = limit[3]

    for j, key in ipairs(KEYS) do
        local ks = key_prefix .. ":" .. key .. ":" .. limit[1]

        local map = redis.call('HMGET', ks, "count", "ts")
        local current = map[1] or max_allowed
        local last_ts = map[2] or now

        local saved = {}

        local fill_rate = max_allowed / duration
        local tokens_to_add = math.floor(fill_rate * (now - last_ts))

        local value = math.min(current + tokens_to_add, max_allowed)

        -- Advance ts only by the amount consumed in refill
        local new_ts = last_ts + (tokens_to_add / fill_rate)

        if value < weight then
            redis.call('HSET', ks, "count", value)
            redis.call('HSET', ks, "ts", new_ts)
            local missing = weight - value
            local elapsed = 0;
            local available_at = missing / fill_rate
            return {0, 0, available_at}
        end

        table.insert(updates, {
            key = ks,
            balance = value - weight,
            ttl = duration
        })
    end
end

-- fill_rate = requests / duration
local remaining = updates[1].balance

-- if request is allowed then only update the count and processed timestamp
for _, update in ipairs(updates) do
    redis.call('HMSET', update.key, "count", update.balance, "ts", now)
    redis.call('EXPIRE', update.key, update.ttl)
    remaining = math.min(remaining, update.balance)
end

return {1, remaining, 0}

