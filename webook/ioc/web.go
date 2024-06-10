package ioc

import (
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/middleware"
	"basic-go/webook/pkg/ginx/middleware/ratelimit"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitGin(mdls []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/login_sms/code/send").
			Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}

}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
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
	})
}
