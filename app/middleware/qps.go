package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/time/rate"
	"inis/app/facade"
	"time"
)

var QoSPoint  = make(map[string]*rate.Limiter)
var QoSGlobal = make(map[string]*rate.Limiter)

// QpsPoint - 单接口限流器
func QpsPoint() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取IP
		ip     := ctx.ClientIP()
		// 获取URL路径
		path   := ctx.Request.URL.Path
		// 获取请求方法
		method := ctx.Request.Method
		// 生成 IP+Path+Method Key
		key    := fmt.Sprintf("ip=%s&path=%s&method=%s", ip, path, method)
		// 从Map中获取对应的访问频率限制器
		limit  := QoSPoint[key]
		// 如果不存在则创建一个新的访问频率限制器
		if limit == nil {
			count := cast.ToInt(facade.AppToml.Get("qps.point", 10))
			limit = rate.NewLimiter(rate.Every(time.Second / 10), count)
			QoSPoint[key] = limit
		}
		// 尝试获取令牌
		if !limit.Allow() {
			ctx.AbortWithStatusJSON(200, gin.H{"code": 429, "msg": facade.Lang(ctx, "请求过于频繁！"), "data": nil})
			return
		}

		ctx.Next()
	}
}

// QpsGlobal - 全局接口限流器
func QpsGlobal() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 获取IP
		ip    := ctx.ClientIP()
		// 从Map中获取对应的访问频率限制器
		limit := QoSGlobal[ip]
		// 如果不存在则创建一个新的访问频率限制器
		if limit == nil {
			count := cast.ToInt(facade.AppToml.Get("qps.global", 50))
			limit = rate.NewLimiter(rate.Every(time.Second / 10), count)
			QoSGlobal[ip] = limit
		}
		// 尝试获取令牌
		if !limit.Allow() {
			ctx.AbortWithStatusJSON(200, gin.H{"code": 429, "msg": facade.Lang(ctx, "请求过于频繁！"), "data": nil})
			return
		}

		ctx.Next()
	}
}