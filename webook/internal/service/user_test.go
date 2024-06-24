package service

import (
	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository"
	repomocks "basic-go/webook/internal/repository/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceLogin(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository
		// ctx      context.Context
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email:    "123@qq.com",
					Password: "$2a$10$.XkMgcHMGX5t.V4eTf2CR.0T7Ru/uh/20kGIDbcdorQ64xMo3Zoym",
					Phone:    "12345678901",
					Ctime:    now,
				}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$.XkMgcHMGX5t.V4eTf2CR.0T7Ru/uh/20kGIDbcdorQ64xMo3Zoym",
				Phone:    "12345678901",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrEmail,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("mock db 错误"))
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email: "123@qq.com",
					// 加密后的 hello#world123
					Password: "$2a$10$.XkMgcHMGX5t.V4eTf2CR.0T7Ru/uh/20kGIDbcdorQ64xMo3Zoym",
					Phone:    "12345678901",
					Ctime:    now,
				}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrEmail,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			u, err := svc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			// 结构体可以直接比较
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello#world123"), bcrypt.DefaultCost)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(res))
}
