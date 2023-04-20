package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"strings"
)

type OAuth struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *OAuth) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"qq":     this.qq,
		"github": this.github,
		"google": this.google,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *OAuth) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"qq":     this.qq,
		"github": this.github,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *OAuth) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"qq": this.qq,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *OAuth) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"qq": this.qq,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *OAuth) INDEX(ctx *gin.Context) {
	this.json(ctx, nil, facade.Lang(ctx, "没什么用！"), 202)
}

func (this *OAuth) qq(ctx *gin.Context) {

	params := this.params(ctx)

	fmt.Println(ctx.Request.Method, params)

	// 登录成功
	if !utils.Is.Empty(params["code"]) {

		item := utils.Curl(utils.CurlRequest{
			Method: "GET",
			Url:    "https://graph.qq.com/oauth2.0/token",
			Query: map[string]string{
				"grant_type":    "authorization_code",
				"client_id":     "102045704",
				"client_secret": "28BAGLMLGHUdijgY",
				"state":         cast.ToString(params["state"]),
				"code":          cast.ToString(params["code"]),
				"redirect_uri":  "https://go.inis.cn/api/auth/qq",
				"fmt":           "json",
			},
		}).Send()

		if item.Error != nil {
			this.json(ctx, nil, "获取token失败："+item.Error.Error(), 500)
			return
		}

		if !utils.Is.Empty(item.Json["access_token"]) {
			unionid := utils.Curl(utils.CurlRequest{
				Url: "https://graph.qq.com/oauth2.0/me",
				Query: map[string]string{
					"access_token": cast.ToString(item.Json["access_token"]),
					"unionid":      "1",
					"fmt":          "json",
				},
			}).Send()

			if !utils.Is.Empty(unionid.Json["openid"]) {
				user := utils.Curl(utils.CurlRequest{
					Url: "https://graph.qq.com/user/get_user_info",
					Query: map[string]string{
						"access_token":       cast.ToString(item.Json["access_token"]),
						"oauth_consumer_key": "102045704",
						"openid":             cast.ToString(unionid.Json["openid"]),
						"fmt":                "json",
					},
				}).Send()

				if !utils.Is.Empty(user.Json["nickname"]) {
					this.json(ctx, user.Json, "成功", 200)
					return
				}
			}
		}

		this.json(ctx, item.Json, "成功", 200)
	}
}

func (this *OAuth) github(ctx *gin.Context) {

	params := this.params(ctx)

	fmt.Println(ctx.Request.Method, params)

	if !utils.Is.Empty(params["code"]) {

		// access_token
		item := utils.Curl(utils.CurlRequest{
			Method: "GET",
			Url:    "https://github.com/login/oauth/access_token",
			Query: map[string]string{
				"client_id":     "Iv1.cc978ca4f5d98345",
				"client_secret": "cca6c49fdeb724f341bae59ccc1e400de12e6c63",
				"code":          cast.ToString(params["code"]),
			},
			Headers: map[string]string{
				"accept": "application/json",
			},
		}).Send()

		if item.Error != nil {
			this.json(ctx, nil, "获取token失败："+item.Error.Error(), 400)
			return
		}

		if !utils.Is.Empty(item.Json["access_token"]) {
			user := utils.Curl(utils.CurlRequest{
				Method: "GET",
				Url:    "https://api.github.com/user",
				Headers: map[string]string{
					"Authorization": "token " + cast.ToString(item.Json["access_token"]),
					"accept":        "application/json",
				},
			}).Send()

			if user.Error != nil {
				this.json(ctx, nil, "获取_token失败："+item.Error.Error(), 400)
				return
			}

			this.json(ctx, user.Json, "登录成功！", 200)
		}
	}
}

func (this *OAuth) google(ctx *gin.Context) {
	clientId := "239457649289-1meff8lsn920gv5i3kug0gef8olgbitg.apps.googleusercontent.com"
	clientKey := "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"
	fmt.Println(clientId, clientKey)
}
