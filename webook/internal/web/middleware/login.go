package middleware

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	// gin 使用 time.Time 类型必须注册编解码，性能会更好
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 如果不需要登录校验 直接返回
		for _, path := range l.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}
		// if ctx.Request.URL.Path == "/users/login" || ctx.Request.URL.Path == "/users/signup" {
		// 	// 直接 return 也会调用 next 的，所以不需要显示调用
		// 	return
		// }
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		// 也可以使用 time.Time 类型
		// now := time.Now().UnixMilli()
		now := time.Now()
		// 刚登录 没有刷新过
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
			return
		}

		// updateTimeVal, _ := updateTime.(int64)
		updateTimeVal, _ := updateTime.(time.Time)
		// if now-updateTimeVal > 60*1000 {
		if now.Sub(updateTimeVal) > time.Second*10 {
			sess.Set("update_time", now)
			sess.Save()
		}
	}
}
