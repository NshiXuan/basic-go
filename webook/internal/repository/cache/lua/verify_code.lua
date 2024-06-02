local key = KEYS[1]
-- 用户输入的 code
local expectedCode = ARGV[1]
local code = redis.call("get", key)
local cntKey = key..":cnt"
-- 转成数字
local cnt = tonumber(redis.call("get", cntKey))
if cnt <= 0 then
  -- 用户一直输入错或者已经用过了，可能有人攻击
  return -1
elseif code == expectedCode then
  -- 验证成功
  -- 用完不能再使用
  -- redis.call("del", key)
  redis.call("del", cntKey)
  return 0
else
  -- 用户手一抖 输入错误
  -- 可验证次数减一
  redis.call("decr ", cntKey)
  return -2
end