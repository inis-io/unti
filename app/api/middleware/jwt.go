package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"inis/app/model"
	"time"
)

func Jwt() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		tokenName := cast.ToString(facade.AppToml.Get("app.token_name", "UNTI_LOGIN_TOKEN"))

		var token string
		if !utils.Is.Empty(ctx.Request.Header.Get("Authorization")) {
			token = ctx.Request.Header.Get("Authorization")
		} else {
			token, _ = ctx.Cookie(tokenName)
		}

		method := []any{"POST", "PUT", "DELETE", "PATCH"}

		if utils.InArray(ctx.Request.Method, method) {

			result := gin.H{"code": 401, "msg": facade.Lang(ctx, "禁止非法操作！"), "data": nil}

			if utils.Is.Empty(token) {
				ctx.JSON(200, result)
				ctx.Abort()
				return
			}

			// 解析token
			jwt := facade.Jwt.Parse(token)
			if jwt.Error != nil {
				result["msg"] = utils.Ternary(jwt.Valid == 0, facade.Lang(ctx, "登录已过期，请重新登录！"), jwt.Error.Error())
				ctx.SetCookie(tokenName, "", -1, "/", "", false, false)
				ctx.JSON(200, result)
				ctx.Abort()
				return
			}

			uid := jwt.Data["uid"]
			cacheName := fmt.Sprintf("user[%v]", uid)

			// 用户缓存不存在 - 从数据库中获取 - 并写入缓存
			if !facade.Cache.Has(cacheName) {

				item := facade.DB.Model(&model.Users{}).Find(uid)
				facade.Cache.Set(cacheName, item, time.Duration(jwt.Valid) * time.Second)
				ctx.Set("user", item)

			} else {

				ctx.Set("user", facade.Cache.Get(cacheName))
			}
		}

		ctx.Next()
	}
}
