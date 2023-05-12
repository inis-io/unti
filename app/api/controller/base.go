package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"hash/fnv"
	"inis/app/facade"
	"inis/app/model"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"
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

	params, ok := ctx.Get("params")

	result = utils.Ternary[map[string]any](ok, cast.ToStringMap(params), make(map[string]any))

	// 合并默认参数
	if empty := utils.Is.Empty(def); !empty {
		for key, val := range def[0] {
			if ok := utils.Is.Empty(result[key]); ok {
				result[key] = val
			}
		}
	}

	return
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

// 从login-token中解析用户信息
func (this base) user(ctx *gin.Context) (result model.Users) {

	// 表数据结构体
	table := model.Users{}
	keys := utils.Struct.Keys(&table)

	if user, ok := ctx.Get("user"); ok {
		for key, val := range cast.ToStringMap(user) {
			if utils.InArray(key, keys) && !utils.Is.Empty(val) {
				utils.Struct.Set(&table, key, val)
			}
		}
	}

	return table
}

// ============================== 分隔线 ==============================

type cache struct{}

// API请求结果是否优先从缓存中获取
func (this cache) enable(ctx *gin.Context) (ok bool) {
	item := cast.ToBool(base{}.param(ctx, "cache", "true"))
	where := item && cast.ToBool(facade.CacheToml.Get("api"))
	return utils.Ternary[bool](where, true, false)
}

// 缓存名称
func (this cache) name(ctx *gin.Context) (name string) {

	// hash 函数
	hash := func(text any) (result string) {
		item := fnv.New32()
		_, err := item.Write([]byte(cast.ToString(text)))
		return cast.ToString(utils.Ternary[any](err != nil, time.Now(), item.Sum32()))
	}

	params, _ := ctx.Get("params")

	body := cast.ToStringMap(params)

	// ========== 此处解决 map 无序问题 - 开始 ==========
	keys := make([]string, 0, len(body))
	for key := range body {
		keys = append(keys, key)
	}
	// 排序 keys
	sort.Strings(keys)
	// ========== 此处解决 map 无序问题 - 开始 ==========

	var buff strings.Builder
	for _, key := range keys {
		buff.WriteString(fmt.Sprintf("%v=%v&", key, body[key]))
	}
	name = buff.String()
	// 去掉最后一个 &
	if !utils.Is.Empty(name) {
		name = name[:len(name)-1]
	}
	// 生产缓存名称
	name = fmt.Sprintf("<%s>%s?hash=%s", ctx.Request.Method, ctx.Request.URL.Path, cast.ToString(hash(name)))

	return
}
