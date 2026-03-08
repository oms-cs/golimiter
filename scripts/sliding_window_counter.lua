-- Parse input arguments
local limits = cjson.decode(ARGV[1])
local now = tonumber(ARGV[2]) / 1000
local key_prefix = "app:rate-limiter:sliding-window-counter:"

-- Initialize tracking variables
local updates = {}
local max_limit = 0

-- Process each rate limit configuration
for i, limit in ipairs(limits) do
    local duration = limit[1]
    local max_allowed = limit[2]
    local weight = limit[3]
    
    max_limit = math.max(max_limit, max_allowed)

    -- Apply rate limiting to each key
    for j, key in ipairs(KEYS) do
        local ks = key_prefix .. key .. ":" .. limit[1]
        local window_id = math.floor(now / duration)

        -- Get current and previous window data
        local prev_window = redis.call('HMGET', ks .. ":prev", "id", "count")
        local current_window = redis.call('HMGET', ks .. ":curr", "id", "count")

        -- Extract window information
        local c_id = tonumber(current_window[1] or 0)
        local c_count = tonumber(current_window[2] or 0)
        local c_ts = c_id * duration
        local p_id = tonumber(prev_window[1] or 0)
        local p_count = tonumber(prev_window[2] or 0)
        local p_ts = p_id * duration

        -- Case 1: Request is in the same window as current
        if window_id == c_id then
            redis.log(redis.LOG_WARNING, "we are in same window as current")
            
            local elapsed_percent = (now - c_ts) / duration
            redis.log(redis.LOG_WARNING, "elapsed_percent" .. elapsed_percent)
            
            local sliding_count = p_count * (1 - elapsed_percent) + c_count
            redis.log(redis.LOG_WARNING, "sliding_count" .. sliding_count)

            -- Check if request exceeds rate limit
            if sliding_count > max_allowed then
                local wait_time = 0
                
                -- If the previous window is still contributing to the count, 
                -- we can calculate the decay time.
                if p_count > 0 then
                    local excess = sliding_count - max_allowed
                    wait_time = (excess / p_count) * duration
                else
                    -- If p_count is 0, the current window is simply full. 
                    -- They must wait for the next window to start.
                    wait_time = duration - (now - c_ts)
                end

                return {0, 0, math.ceil(wait_time)}
            end

            -- Queue update for current window
            table.insert(updates, {
                key = ks .. ":curr",
                id = window_id,
                count = c_count + weight,
                ttl = duration * 2,
                remaining = math.max(0, max_allowed - (sliding_count + weight)),
                is_current = true
            })

        -- Case 2: Request is in the next window (boundary case)
        elseif window_id == c_id + 1 then
            -- Check sliding count at the boundary
            local elapsed_percent = (now - c_ts) / duration
            local sliding_count = math.ceil(c_count * (1 - elapsed_percent))

            -- Check if request exceeds rate limit at boundary
            if sliding_count + weight > max_allowed then
                local wait_time = 0
                
                if c_count > 0 then
                    local excess = sliding_count + weight - max_allowed
                    wait_time = (excess / c_count) * duration
                else
                    wait_time = duration - (now - c_ts)
                end
                
                return {0, 0, math.ceil(wait_time)}
            end

            -- Queue updates: move current to previous, start new current
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

        -- Case 3: Request is in a completely new window
        else
            redis.log(redis.LOG_WARNING, "we are in both new windows")
            
            -- Check if single request exceeds limit
            if weight >= max_allowed then
                return {0, 0, duration}
            end

            -- Queue updates: reset both windows
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

-- Initialize remaining requests to the maximum limit
local remaining = max_limit

-- Apply all queued updates if request is allowed
for _, update in ipairs(updates) do
    redis.call('HSET', update.key, "id", update.id, "count", update.count)
    redis.call('EXPIRE', update.key, update.ttl)
    
    -- Track remaining requests from current windows
    if update.is_current then
        remaining = math.max(0, math.min(remaining, update.remaining))
    end
end

-- Return success: allowed, remaining requests, wait time
return {1, remaining, 0}
