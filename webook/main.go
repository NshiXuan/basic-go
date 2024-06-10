package main

import (
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/service/sms/memory"
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/middleware"
	"basic-go/webook/pkg/ginx/middleware/ratelimit"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	// db := initDB()
	// rdb := initRedisDB()
	// server := initWebServer(rdb)
	// u := initUser(db, rdb)
	// u.RegisterRoutes(server)
	server := InitWebServer()
	server.Run(":8080")
}

func initWebServer(rdb redis.Cmdable) *gin.Engine {
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		println("第一个 middleware")
	})

	server.Use(ratelimit.NewBuilder(rdb, time.Second, 100).Build())

	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://localhost:8088"},
		// AllowMethods:     []string{"PUT", "PATCH"},
		// AllowHeaders:  []string{"Origin"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 不加 ExposeHeaders ，前端获取不到 token
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			fmt.Printf("origin: %v\n", origin)
			if strings.HasPrefix(origin, "http://localhost") || strings.Contains(origin, "chrome-extension") {
				fmt.Printf("has origin: %v\n", origin)
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("cJ5rC2kQ4dF5oN3dH3wG4jT6bO4rU1dS"), []byte("uX7lE7bW8qM8tE4yN6dR1uD7cA4eD2tW"))
	// 参数：最大空闲连接数，tcp，连接信息，密码，加密key，key
	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("cJ5rC2kQ4dF5oN3dH3wG4jT6bO4rU1dS"))
	// if err != nil {
	// 	panic(err)
	// }
	server.Use(sessions.Sessions("mysession", store))
	// server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").Build())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/login").IgnorePaths("/users/signup").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/users/login_sms/code/send").
		Build())
	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDao(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	svc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCoderService(codeRepo, smsSvc)

	u := web.NewUserHandler(svc, codeSvc)
	return u
}
