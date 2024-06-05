//go:build wireinject

// 让 wire 来注入这里的代码
package wire

import (
	"basic-go/wire/repository"
	"basic-go/wire/repository/dao"

	"github.com/google/wire"
)

func InitReposity() *repository.UserRepository {
	// Build 传入各个组件的初始化方法, 会帮我们构造、编排顺序
	wire.Build(repository.NewUserRepository, dao.NewUserDao, InitDB)
	return new(repository.UserRepository)
}
