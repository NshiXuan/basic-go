//go:build wireinject

package main

import (
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	"basic-go/webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB,
		ioc.InitRedis,

		// 初始化 DAO
		dao.NewUserDao,

		// 初始化缓存
		cache.NewUserCache,
		cache.NewCodeCache,

		// 初始化业务
		repository.NewCacheUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		ioc.InitSMSService,

		// 初始化 handler
		web.NewUserHandler,

		// 中间件、注册路由呢？
		// gin.Default,
		ioc.InitGin,
		ioc.InitGinMiddlewares,
	)
	return new(gin.Engine)
}
