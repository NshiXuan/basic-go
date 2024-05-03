package middleware

import (
	"basic-go/webook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 如果不需要登录校验 直接返回
		for _, path := range l.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}

		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			// 没登录 有人瞎搞
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		claims := &web.UserClaims{}
		// ParseWithClaims 一定要传递指针, 类似 json 解析
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("cJ5rC2kQ4dF5oN3dH3wG4jT6bO4rU1dS"), nil
		})
		if err != nil {
			// 瞎写，Bearer xxx123 ，默认为没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 如果 token 过期了, Valid 为 false
		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 换设备不会触发下面的判断 触发条件一般是有人复制了 token 到其它端使用,比如: 浏览器的 token 到手机上使用
		// 用户浏览器升级也会触发
		if claims.UserAgent != ctx.Request.UserAgent() {
			log.Println("UserAgent 不一致")
			// 严重的安全问题
			// 加监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新登录态
		now := time.Now()
		// login 设置的 token 过期实际为 1 min , 所以 < 50 时代表过了 10s
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("cJ5rC2kQ4dF5oN3dH3wG4jT6bO4rU1dS"))
			if err != nil {
				log.Println("jwt 续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}
		ctx.Set("claims", claims)
		// ctx.Set("userId", claims.Uid)
	}
}
