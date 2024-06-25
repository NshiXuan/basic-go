package repository

import (
	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/cache"
	cachemocks "basic-go/webook/internal/repository/cache/mocks"
	"basic-go/webook/internal/repository/dao"
	daomocks "basic-go/webook/internal/repository/dao/mocks"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedUserRepositoryFindById(t *testing.T) {
	now := time.Now()
	fmt.Printf("now: %v\n", now)
	// 去掉毫秒以外的部分
	now = time.UnixMilli(now.UnixMilli())
	fmt.Printf("now after: %v\n", now)
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache)

		// FindById参数
		ctx context.Context
		id  int64

		// FindById返回值
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，但是查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				// 缓存未命中 查了缓存 没有结果
				c := cachemocks.NewMockUserCache(ctrl)
				// 预期缓存调用 Get 方法，返回空的 User
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotFound)

				d := daomocks.NewMockUserDao(ctrl)
				// 预期从数据库中获取到一个 User 返回
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "password",
					Phone: sql.NullString{
						String: "15212345678",
						Valid:  true,
					},
					Ctime: now.UnixMilli(),
					Utime: now.UnixMilli(),
				}, nil)

				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "password",
					Phone:    "15212345678",
					Ctime:    now,
				}).Return(nil)

				return d, c
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "password",
				Phone:    "15212345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				// 缓存未命中 查了缓存 没有结果
				c := cachemocks.NewMockUserCache(ctrl)
				// 预期缓存调用 Get 方法，返回空的 User
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "password",
					Phone:    "15212345678",
					Ctime:    now,
				}, nil)

				d := daomocks.NewMockUserDao(ctrl)
				return d, c
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "password",
				Phone:    "15212345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，但是查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				// 缓存未命中 查了缓存 没有结果
				c := cachemocks.NewMockUserCache(ctrl)
				// 预期缓存调用 Get 方法，返回空的 User
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotFound)

				d := daomocks.NewMockUserDao(ctrl)
				// 预期从数据库中获取到一个 User 返回
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{}, errors.New("mock db 错误"))
				return d, c
			},
			ctx:      context.Background(),
			id:       123,
			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
			// Set 为并发，所以需要 Sleep
			time.Sleep(time.Second)
		})
	}
}
