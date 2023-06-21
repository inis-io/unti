package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"strings"
)

type Proxy struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *Proxy) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *Proxy) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *Proxy) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *Proxy) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *Proxy) INDEX(ctx *gin.Context) {

	params := this.params(ctx)
	filter := []string{"i-url", "i-type"}

	// // 创建 CookieJar
	// jar, _ := cookiejar.New(nil)
	//
	// client := &http.Client{
	// 	Jar: jar,
	// }
	//
	// // cookie 数据不为空
	// if !utils.Is.Empty(ctx.Request.Cookies()) {
	// 	client.Jar.SetCookies(ctx.Request.URL, ctx.Request.Cookies())
	// }

	item := utils.Curl(utils.CurlRequest{
		// Client:  client,
		Url:     cast.ToString(params["i-url"]),
		Method:  strings.ToUpper(ctx.Request.Method),
		Headers: this.headers(ctx),
		Data: 	 this.data(ctx, filter),
		Query:   this.query(ctx, filter),
	}).Send()

	if item.Error != nil {
		ctx.JSON(500, item.Error.Error())
		return
	}

	switch strings.ToUpper(cast.ToString(params["i-type"])) {
	case "JSON":
		ctx.JSON(200, item.Json)
	case "TEXT":
		ctx.JSON(200, item.Text)
	case "BYTE":
		ctx.JSON(200, item.Byte)
	default:
		ctx.JSON(200, item.Json)
	}
}

// query - 获取 query 位置的参数
func (this *Proxy) query(ctx *gin.Context, filter []string) (result map[string]any) {

	result = make(map[string]any)

	// query 数据不为空
	if !utils.Is.Empty(ctx.Request.URL.Query()) {
		item := utils.Parse.Params(utils.Parse.ParamsBefore(ctx.Request.URL.Query()))
		// 合并 params 参数
		for key, val := range item {
			if !utils.InArray(key, filter) {
				result[key] = val
			}
		}
	}

	return
}

// data - 获取 body 位置的参数
func (this *Proxy) data(ctx *gin.Context, filter []string) (result map[string]any) {

	result = make(map[string]any)

	// 只获取 key
	for key := range this.query(ctx, filter) {
		filter = append(filter, key)
	}

	for key, val := range this.params(ctx) {
		if !utils.InArray(key, filter) {
			result[key] = val
		}
	}

	return
}