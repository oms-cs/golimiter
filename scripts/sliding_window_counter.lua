-- Parse input arguments
local limits = cjson.decode(ARGV[1])
local now = tonumber(ARGV[2]) / 1000
local key_prefix = "app:rate-limiter:sliding-window-counter:"

local updates = {}
local max_limit = 0

for i, limit in ipairs(limits) do
    local duration = limit[1]
    local max_allowed = limit[2]
    local weight = limit[3]
    
    max_limit = math.max(max_limit, max_allowed)

    for j, key in ipairs(KEYS) do
        local ks = key_prefix .. key .. ":" .. duration
        local window_id = math.floor(now / duration)

        local prev_window = redis.call('HMGET', ks .. ":prev", "id", "count")
        local current_window = redis.call('HMGET', ks .. ":curr", "id", "count")

        local c_id = tonumber(current_window[1] or 0)
        local c_count = tonumber(current_window[2] or 0)
        local p_id = tonumber(prev_window[1] or 0)
        local p_count = tonumber(prev_window[2] or 0)

        -- Case 1: We are in the same window as the currently tracked window
        if window_id == c_id then
            local c_ts = c_id * duration
            local elapsed_percent = (now - c_ts) / duration
            
            -- Only use previous count if it's actually the immediate prior window
            local valid_p_count = (p_id == window_id - 1) and p_count or 0
            local sliding_count = valid_p_count * (1 - elapsed_percent) + c_count

            -- BUGFIX: Add weight to the check
            if sliding_count + weight > max_allowed then
                local wait_time = 0
                if valid_p_count > 0 then
                    local excess = (sliding_count + weight) - max_allowed
                    wait_time = (excess / valid_p_count) * duration
                else
                    wait_time = duration - (now - c_ts)
                end
                return {0, 0, math.ceil(wait_time)}
            end

            table.insert(updates, {
                key = ks .. ":curr",
                id = window_id,
                count = c_count + weight,
                ttl = duration * 2,
                remaining = math.max(0, max_allowed - (sliding_count + weight)),
                is_current = true
            })

        -- Case 2: We have transitioned exactly one window forward
        elseif window_id == c_id + 1 then
            -- BUGFIX: Calculate elapsed percent using the start of the NEW window
            local new_window_ts = window_id * duration
            local elapsed_percent = (now - new_window_ts) / duration
            
            -- The old current becomes the new previous
            local sliding_count = math.ceil(c_count * (1 - elapsed_percent))

            if sliding_count + weight > max_allowed then
                local wait_time = 0
                if c_count > 0 then
                    local excess = (sliding_count + weight) - max_allowed
                    wait_time = (excess / c_count) * duration
                else
                    wait_time = duration - (now - new_window_ts)
                end
                return {0, 0, math.ceil(wait_time)}
            end

            table.insert(updates, {
                key = ks .. ":prev",
                id = c_id,
                count = c_count,
                ttl = duration * 2,
                is_current = false,
                remaining = max_allowed
            })
            
            table.insert(updates, {
                key = ks .. ":curr",
                id = window_id,
                count = weight,
                ttl = duration * 2,
                is_current = true,
                remaining = max_allowed - (sliding_count + weight)
            })

        -- Case 3: We skipped windows entirely (or first run)
        else
            if weight > max_allowed then
                -- Return -1 to signify the payload is fundamentally too large
                return {0, 0, -1} 
            end

            table.insert(updates, {
                key = ks .. ":prev",
                id = window_id - 1,
                count = 0,
                ttl = duration,
                is_current = false,
                remaining = max_allowed
            })
            
            table.insert(updates, {
                key = ks .. ":curr",
                id = window_id,
                count = weight,
                ttl = duration * 2,
                is_current = true,
                remaining = max_allowed - weight
            })
        end
    end
end

local remaining = max_limit

for _, update in ipairs(updates) do
    redis.call('HSET', update.key, "id", update.id, "count", update.count)
    redis.call('EXPIRE', update.key, update.ttl)
    
    if update.is_current then
        remaining = math.max(0, math.min(remaining, update.remaining))
    end
end

return {1, remaining, 0}