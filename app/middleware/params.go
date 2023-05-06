package middleware

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func Params() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 判断是否在数组中
		inArray := func(value any, array []any) bool {
			for _, val := range array {
				if val == value {
					return true
				}
			}
			return false
		}

		method := ctx.Request.Method
		params := make(map[string]any)

		// 拷贝一份 body
		body, _ := io.ReadAll(ctx.Request.Body)
		// ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))

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

		if inArray(method, []any{"POST", "PUT", "DELETE", "PATCH"}) {
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
