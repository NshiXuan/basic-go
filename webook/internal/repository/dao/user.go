package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDao interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	Insert(ctx context.Context, u User) error
}

type GORMUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GORMUserDao{
		db: db,
	}
}

func (dao *GORMUserDao) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	// NOTE(nsx): WithContext 让 db 保持链路
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConfilictErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConfilictErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *GORMUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	// 不用检查，找不到就返回
	err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (dao *GORMUserDao) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	// 不用检查，找不到就返回
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	// 不用检查，找不到就返回
	err := dao.db.WithContext(ctx).First(&u, "phone = ?", phone).Error
	return u, err
}

// User 直接对应数据库表
// 有些人叫 entity 或 model 或 PO (persistent object)
type User struct {
	Id       int64
	Email    sql.NullString `gorm:"primaryKey,autoIncrement"`
	Password string         `gorm:"unique"`
	// 不能直接使用 string 类型，因为如果没有传 phone ，就有很多为空字符串的情况，这时候就不是唯一索引了，可以使用 sql.NullString
	// 唯一索引，可以有多个空值，但是不能有多个空字符串
	Phone sql.NullString `gorm:"unique"`

	// 毫秒
	Ctime int64
	Utime int64
}
