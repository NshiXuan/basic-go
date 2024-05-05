-- 你的验证码在 Redis 上的 key
-- phone_code:login:180xxx
local key = KEYS[1]
-- 验证次数，最多重复三次，这个记录还可以验证几次
-- phone_code:login:180xxx:cnt
local cntKey = key..":cnt"
-- 你的验证码
local val = ARGV[1]
-- 过期时间
local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    -- 如果 key 存在，但是没用过期时间
    -- 系统错误，手动设置了 key , 但是没有设置过期时间
    return -2
-- 540 = 600 -60 九分钟
elseif ttl == -2 or ttl < 540 then
  redis.call("set", key, val)
  redis.call("expire", key, 600)
  redis.call("set", cntKey, val)
  redis.call("expire", cntKey, 600)
  -- 符合预期
  return 0
else 
  -- 发送太频繁
  return -1
end