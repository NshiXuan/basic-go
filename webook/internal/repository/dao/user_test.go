package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGORMUserDaoInsert(t *testing.T) {
	tests := []struct {
		name string
		mock func(t *testing.T) *sql.DB

		// 输入
		ctx context.Context
		u   User

		// 输出
		wantErr error
		wantId  int64
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				// 增删改使用 ExpectExec ，查使用 ExpectQuery
				// * 代表正则表达式，只要符合预期 INSERT INTO `users` 语句就可以
				// sqlmock.NewResult 参数是插入的 id ，影响的行数
				mock.ExpectExec("INSERT INTO `users`.*").WillReturnResult(sqlmock.NewResult(3, 1))
				return mockDB
			},
			u: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantId: 3,
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users`.*").WillReturnError(&mysql.MySQLError{
					Number: 1062,
				})
				return mockDB
			},
			u:       User{},
			wantErr: ErrUserDuplicateEmail,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users`.*").WillReturnError(errors.New("数据库错误"))
				return mockDB
			},
			u:       User{},
			wantErr: errors.New("数据库错误"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: tt.mock(t),
				// 跳过 gorm 初始化的时候调用 show version 的步骤
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// mock DB 不需要 ping
				DisableAutomaticPing: true,
				// GORM 默认会开启一个事务来执行 sql 语句，测试的时候跳过
				SkipDefaultTransaction: true,
			})
			d := NewUserDao(db)
			err = d.Insert(tt.ctx, tt.u)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
