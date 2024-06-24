package web

import (
	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/service"
	svcmocks "basic-go/webook/internal/service/mocks"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestEncrypt(t *testing.T) {
	password := "hello#world123"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("encrypted: %v\n", string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestMock(t *testing.T) {
	// 1.初始化控制器
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	usersvc := svcmocks.NewMockUserService(ctrl)

	// 2.EXPECT 设置预期调用, 预期参数与预期返回值
	// gomock.Any() 表示随便传递什么参数
	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))

	// 3.发起调用, 参数要与上面预期的参数一致，不然报错
	// 且调用次数要与预期的此时一致，不然报错
	// 返回值为预期的返回值 mock error
	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}

func TestUserHandlerSignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数不对，Bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
	"email": "123@qq.com",
	"password": "hello#world123"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
	{
	"email": "123@q",
	"password": "hello#world123"
	}
`,
			wantCode: http.StatusOK,
			wantBody: "非法邮箱格式",
		},
		{
			name: "两次输入密码不匹配",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world1234",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "两次输入密码不对",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello123",
	"confirmPassword": "hello123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "密码必须包含字母、数字、特殊字符，并且不少于八位",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrUserDuplicateEmail)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("随便一个error"))
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// httptest.NewRecorder 实现了 io.Write 接口，用于存储 http 响应的数据，我们可以从中获取数据进行校验
			resp := httptest.NewRecorder()

			// ServeHTTP 是 HTTP 请求进去 gin 框架人入口，响应回写到 resp 中
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}
