package controller

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type base struct {
	cache
}

type ApiInterface interface {
	IGET(ctx *gin.Context)
	IPUT(ctx *gin.Context)
	IDEL(ctx *gin.Context)
	IPOST(ctx *gin.Context)
	INDEX(ctx *gin.Context)
	// IPATCH(ctx *gin.Context)
}

func (this base) json(ctx *gin.Context, data, msg, code any) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": cast.ToInt(code),
		"msg":  cast.ToString(msg),
		"data": data,
	})
}

// Call 方法调用 - 资源路由本体
func (this base) call(allow map[string]any, name string, params ...any) (err error) {

	// 判断 allow 是否为空
	if empty := utils.Is.Empty(allow); empty {
		return errors.New("allow is empty")
	}

	// 判断 name 是否在 allow 中
	if _, ok := allow[name]; !ok {
		return errors.New(name + " is not in allow")
	}

	method := reflect.ValueOf(allow[name])

	if len(params) != method.Type().NumIn() {
		return errors.New("输入参数的数量不匹配！")
	}

	in := make([]reflect.Value, len(params))

	for key, val := range params {
		in[key] = reflect.ValueOf(val)
	}

	method.Call(in)

	return nil
}

// 获取单个参数
func (this base) param(ctx *gin.Context, key string, def ...any) any {

	var value map[string]any

	if empty := utils.Is.Empty(def); !empty {
		value = map[string]any{key: def[0]}
	} else {
		value = map[string]any{key: nil}
	}

	params := this.params(ctx, value)

	return params[key]
}

// params 获取全部参数 , def map[string]any
func (this base) params(ctx *gin.Context, def ...map[string]any) (result map[string]any) {

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

	// 合并默认参数
	if empty := utils.Is.Empty(def); !empty {
		for key, val := range def[0] {
			if ok := utils.Is.Empty(params[key]); ok {
				params[key] = val
			}
		}
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	return params
}

// 获取单个请求头信息
func (this base) header(ctx *gin.Context, key string, def ...any) (result string) {
	result = ctx.GetHeader(key)
	if empty := utils.Is.Empty(result); empty {
		if !utils.Is.Empty(def) {
			result = def[0].(string)
		}
	}
	return
}

// 获取全部请求头信息
func (this base) headers(ctx *gin.Context) (result map[string]any) {
	result = make(map[string]any)
	for key, val := range ctx.Request.Header {
		result[key] = val[0]
	}
	return
}

// ============================== 分隔线 ==============================

type cache struct{}

// API请求结果是否优先从缓存中获取
func (this cache) enable(ctx *gin.Context) (ok bool) {
	item  := cast.ToBool(base{}.param(ctx, "cache", "true"))
	where := item && cast.ToBool(facade.CacheToml.Get("api"))
	return utils.Ternary[bool](where, true, false)
}

// 缓存名称
func (this cache) name(ctx *gin.Context) (name string) {
	params := base{}.params(ctx)
	name = fmt.Sprintf("<%s>%s?%s", ctx.Request.Method, ctx.Request.URL.Path, utils.Format.Query(params))
	return
}
