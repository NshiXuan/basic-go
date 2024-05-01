package main

import (
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/middleware"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	u := initUser(db)
	u.RegisterRoutes(server)
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		println("第一个 middleware")
	})
	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://localhost:8088"},
		// AllowMethods:     []string{"PUT", "PATCH"},
		// AllowHeaders:  []string{"Origin"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
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
	// store := memstore.NewStore([]byte("cJ5rC2kQ4dF5oN3dH3wG4jT6bO4rU1dS"), []byte("uX7lE7bW8qM8tE4yN6dR1uD7cA4eD2tW"))
	// 参数：最大空闲连接数，tcp，连接信息，密码，加密key，key
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("cJ5rC2kQ4dF5oN3dH3wG4jT6bO4rU1dS"))
	if err != nil {
		panic(err)
	}
	server.Use(sessions.Sessions("mysession", store))
	server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").Build())
	return server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err)
	}
	if err := dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDao(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}
