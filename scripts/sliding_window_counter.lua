
local limits = cjson.decode(ARGV[1])

local now = cjson.decode(ARGV[3]) / 1000

local key_prefix = "app:rate-limiter:sliding-window-counter:"

local updates = {}

for i, limit in ipairs(limits) do
    local duration = limit[1]
    local max_allowed = limit[2]
    local weight = limit[3]

    for j, key in ipairs(KEYS) do
        local ks = key_prefix .. ":" .. key .. ":" .. limit[1]

        local window_id = math.ceil(now / duration)

        local prev_window = redis.call('HMGET', ks .. ":prev", "id", "count")
        local current_window = redis.call('HMGET', ks .. ":curr", "id", "count")

        --window info
        local c_id = tonumber(current_window[1] or 0)
        local c_count = tonumber(current_window[2] or 0)
        local c_ts = c_id * duration
        local p_id = tonumber(prev_window[1] or 0)
        local p_count = tonumber(prev_window[2] or 0)
        local p_ts = p_id * duration

        -- check if window id matches current window
        if window_id == c_id then
            local elapsed_percent = (now - p_ts) / duration
            local sliding_count = p_count * (1 - elapsed_percent) + c_count
            if sliding_count > max_allowed then
                return 0
            end
            updates[ks .. ":curr"] = {id = window_id, count = sliding_count + weight, ttl = duration*2}

        elseif window_id == c_id + 1 then
            updates[ks .. ":prev"] = {id = c_id, count = c_count, ttl = duration*2}
            updates[ks .. ":curr"] = {id = window_id, count = weight, ttl = duration*2}
        
        else
            if (weight) > max_allowed then return 0 end
            updates[ks.. ":prev"] = {id = window_id - 1, count = 0, ttl = duration}
            updates[ks .. ":curr"] = {id = window_id, count = weight, ttl = duration*2}
        end
    end
end

-- if request is allowed then only update the count and processed timestamp
for _, update in ipairs(updates) do
    redis.call('HMSET', update.key, "id", update.id, "count", update.count)
    redis.call('EXPIRE', update.key, update.ttl)
end

return 1