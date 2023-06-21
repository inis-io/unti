package middleware

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func Params() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 挂载域名信息
		go domain(ctx)
		// 挂载IP信息
		go clientIP(ctx)
		// 挂载端口号
		go port(ctx)

		method := ctx.Request.Method
		params := make(map[string]any)

		// 拷贝一份 body
		body, _ := io.ReadAll(ctx.Request.Body)

		content := map[string]any{
			"type":  ctx.GetHeader("Content-Type"),
			"body":  string(body),
			"form":  ctx.Request.Form,
			"query": ctx.Request.URL.Query(),
		}

		// 处理空 Content-Type
		if utils.Is.Empty(content["type"]) {

			if method == "GET" || method == "DELETE" {

				content["type"] = "application/x-www-form-urlencoded"

			} else {

				if !utils.Is.Empty(content["body"]) {
					content["type"] = "application/json"
				} else if !utils.Is.Empty(content["query"]) && !utils.Is.Empty(content["form"]) {
					content["type"] = "application/x-www-form-urlencoded"
				}
			}
		}

		if utils.In.Array(method, []any{"POST", "PUT", "DELETE", "PATCH"}) {
			if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
				if !errors.Is(err, http.ErrNotMultipart) {
					// fmt.Println(err)
				}
			}
		}

		contentType := cast.ToString(content["type"])

		// body 数据不为空
		if !utils.Is.Empty(content["body"]) {

			var item map[string]any

			switch {
			// 解析 application/json 数据
			case strings.Contains(contentType, "application/json"):

				// 合并 params 参数
				item = cast.ToStringMap(utils.Json.Decode(string(body)))
				// 合并 params 参数
				if !utils.Is.Empty(item) {
					for key, val := range item {
						params[key] = val
					}
				}

			// 解析 application/x-www-form-urlencoded 数据
			case strings.Contains(contentType, "application/x-www-form-urlencoded"):

				// 我的body数据长这样id=2&domian=inis，我需要把它转成url.Values类型
				values, err := url.ParseQuery(string(body))
				if err != nil {
					break
				}
				item = utils.Parse.Params(utils.Parse.ParamsBefore(values))
				// 合并 params 参数
				if !utils.Is.Empty(item) {
					for key, val := range item {
						params[key] = val
					}
				}

			// 解析 multipart/form-data 数据
			case strings.Contains(contentType, "multipart/form-data"):

				// 将字符串转换为 bytes.Buffer 类型
				bodyBuffer := bytes.NewBufferString(string(body))
				boundary := strings.Split(contentType, "boundary=")

				if len(boundary) != 2 {
					break
				}

				// 创建 multipart.Reader 对象，用于解析 multipart/form-data 格式的请求体
				multipartReader := multipart.NewReader(bodyBuffer, boundary[1])

				// 解析 multipart/form-data 格式的请求体
				formData, err := multipartReader.ReadForm(0)
				if err != nil {
					break
				}

				// 将表单项转换为 url.Values 类型
				values := url.Values{}
				for key, valuesList := range formData.Value {
					for _, value := range valuesList {
						values.Add(key, value)
					}
				}

				item = utils.Parse.Params(utils.Parse.ParamsBefore(values))

				// 合并 params 参数
				if !utils.Is.Empty(item) {
					for key, val := range item {
						params[key] = val
					}
				}
				// 解析 application/xml 数据
				// 解析 text/plain 数据
				// 解析 text/xml 数据
			}
		}

		// query 数据不为空
		if !utils.Is.Empty(content["query"]) {

			item := utils.Parse.Params(utils.Parse.ParamsBefore(ctx.Request.URL.Query()))
			// 合并 params 参数
			for key, val := range item {
				params[key] = val
			}
		}
		ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))

		ctx.Set("params", params)
	}
}

// 获取域名
func domain(ctx *gin.Context) (result string) {

	host := ctx.Request.Header.Get("X-Host")
	host = utils.Ternary[string](utils.Is.Empty(host), ctx.Request.Host, host)

	// 得到 主机地址 和 端口号
	info := []string{"localhost", "80"}
	if strings.Contains(host, ":") {
		info = strings.Split(host, ":")
	} else {
		info[0] = host
	}

	// 得到 SSL 协议
	scheme := ctx.Request.Header.Get("X-Scheme")
	if utils.Is.Empty(scheme) {
		scheme = utils.Ternary[string](cast.ToInt(info[1]) == 443, "https", "http")
	}

	// 组装域名
	result = scheme + "://" + info[0]
	if !utils.InArray[int](cast.ToInt(info[1]), []int{80, 443}) {
		result += ":" + info[1]
	}

	// 存储到缓存中
	go func() {
		if cast.ToBool(facade.CacheToml.Get("open")) {
			facade.Cache.Set("domain", result, 0)
		}
		// 存储到上下文中
		ctx.Set("domain", result)
	}()

	return result
}

// 获取客户端IP
func clientIP(ctx *gin.Context) (result string) {

	// 获取IP
	result = ctx.Request.Header.Get("X-Real-IP")
	if utils.Is.Empty(result) {
		result = ctx.Request.Header.Get("X-Forwarded-For")
	}
	if utils.Is.Empty(result) {
		result = ctx.ClientIP()
	}

	// 存储到上下文中
	ctx.Set("ip", result)

	return result
}

// 获取端口号
func port(ctx *gin.Context) (result int) {

	host := ctx.Request.Header.Get("X-Host")
	host = utils.Ternary[string](utils.Is.Empty(host), ctx.Request.Host, host)

	// 得到 主机地址 和 端口号
	if strings.Contains(host, ":") {
		info := strings.Split(host, ":")
		result = cast.ToInt(info[1])
	}

	if utils.Is.Empty(result) {
		result = utils.Ternary[int](ctx.Request.Header.Get("X-Scheme") == "https", 443, 80)
	}

	// 存储到上下文中
	ctx.Set("port", result)

	return result
}