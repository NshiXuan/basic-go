package integration

import (
	"basic-go/webook/internal/web"
	"basic-go/webook/ioc"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string

		// 需要考虑准备的数据
		before func(t *testing.T)
		// 验证的数据，数据库的数据对不对，Redis 的数据对不对
		after    func(t *testing.T)
		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
				// 不需要 也就是 Redis 什么也没有
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 需要清理的数据
				val, err := rdb.GetDel(ctx, "phone_code:login:15212345678").Result()
				cancel()
				assert.NoError(t, err)
				// 验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
			{
				"phone": "15212345678"
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 先设置验证码
				_, err := rdb.Set(ctx, "phone_code:login:15212345678", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 需要清理的数据
				val, err := rdb.GetDel(ctx, "phone_code:login:15212345678").Result()
				cancel()
				assert.NoError(t, err)
				// 验证码还是 123456 没有被覆盖
				assert.Equal(t, "123456", val)
			},
			reqBody: `
			{
				"phone": "15212345678"
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送太频繁,请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 这个手机号已经有了一个验证码，但是没有过期时间
				_, err := rdb.Set(ctx, "phone_code:login:15212345678", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 需要清理的数据
				val, err := rdb.GetDel(ctx, "phone_code:login:15212345678").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			reqBody: `
			{
				"phone": "15212345678"
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "手机号码为空",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
			{
				"phone": ""
			}
			`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "输入有误",
			},
		},
		{
			name: "数据格式错误",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
			{
				"phone": 
			}
			`,
			wantCode: 400,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// httptest.NewRecorder 实现了 io.Write 接口，用于存储 http 响应的数据，我们可以从中获取数据进行校验
			resp := httptest.NewRecorder()

			// ServeHTTP 是 HTTP 请求进去 gin 框架人入口，响应回写到 resp 中
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}

			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)

			tc.after(t)
		})
	}
}
