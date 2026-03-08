local limits = cjson.decode(ARGV[1])
local now_ms = tonumber(ARGV[2])
local now_sec = now_ms / 1000
local weight = tonumber(ARGV[3] or 1)

local longest_duration = 0
local updates = {}
local global_min_remaining = 999999999
local key_prefix = "app:rate-limiter:sliding-window-log:"


for i, limit in ipairs(limits) do
    local duration = tonumber(limit[1])
    local max_allowed = tonumber(limit[2])
    local precision = tonumber(limit[3] or duration)
    redis.log(redis.LOG_WARNING, "Duration: " .. duration .. " Max allowed: " .. max_allowed .. " Precision: " .. precision)
    
    longest_duration = math.max(longest_duration, duration)
    precision = math.min(precision, duration)
    local blocks = math.ceil(duration / precision)

    local block_id = math.floor(now_sec / precision)
    local trim_before = block_id - blocks + 1
    local count_key_base = duration .. ':' .. precision .. ':'
    local ts_key = count_key_base .. 'o'

    for j, key in ipairs(KEYS) do
        local ks = key_prefix .. key
        local old_ts = tonumber(redis.call('HGET', ks, ts_key) or trim_before)
        
        -- 1. CLEANUP EXPIRED BLOCKS
        local decr = 0
        local dele = {}
        local trim_to = math.min(trim_before, old_ts + blocks)

        for old_block = old_ts, trim_to - 1 do
            local bkey = count_key_base .. old_block
            local bcount = redis.call('HGET', ks, bkey)
            if bcount then
                decr = decr + tonumber(bcount)
                table.insert(dele, bkey)
            end
        end

        local current_total = 0
        if #dele > 0 then
            redis.call('HDEL', ks, unpack(dele))
            current_total = tonumber(redis.call('HINCRBY', ks, count_key_base, -decr))
        else
            current_total = tonumber(redis.call('HGET', ks, count_key_base) or 0)
        end

        -- 2. ACCURATE WAIT TIME CALCULATION
        if current_total + weight > max_allowed then
            local needed = (current_total + weight) - max_allowed
            local found_decr = 0
            local wait_time = 0

            -- Look at active blocks starting from the oldest
            for b = trim_before, block_id do
                local bkey = count_key_base .. b
                local bcount = tonumber(redis.call('HGET', ks, bkey) or 0)
                
                if bcount > 0 then
                    found_decr = found_decr + bcount
                    if found_decr >= needed then
                        -- The wait time is when THIS block falls out of the window
                        -- A block 'b' falls out at: (b + blocks) * precision
                        local expiry_ts = (b + blocks) * precision
                        wait_time = math.max(0, expiry_ts - now_sec)
                        break
                    end
                end
            end
            
            -- If we still haven't found enough, wait for the full duration
            if wait_time == 0 then wait_time = precision end
            redis.log(redis.LOG_WARNING, "Wait time: " .. wait_time)
            return {0, 0, math.ceil(wait_time)}
        end

        table.insert(updates, {
            key = ks,
            ts_key = ts_key,
            trim_before = trim_before,
            count_key_base = count_key_base,
            block_id = block_id,
            remaining = max_allowed - (current_total + weight)
        })
        
        global_min_remaining = math.min(global_min_remaining, max_allowed - (current_total + weight))
    end
end

-- 3. COMMIT
for _, u in ipairs(updates) do
    redis.call('HSET', u.key, u.ts_key, u.trim_before)
    redis.call('HINCRBY', u.key, u.count_key_base, weight)
    redis.call('HINCRBY', u.key, u.count_key_base .. u.block_id, weight)
    redis.call('EXPIRE', u.key, math.ceil(longest_duration))
end

return {1, global_min_remaining, 0}