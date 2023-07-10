package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"inis/app/model"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

type base struct {
	meta
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

// params 获取全部参数
func (this base) params(ctx *gin.Context, def ...map[string]any) (result map[string]any) {

	params, ok := ctx.Get("params")

	result = utils.Ternary[map[string]any](ok, cast.ToStringMap(params), make(map[string]any))

	// 合并默认参数
	if empty := utils.Is.Empty(def); !empty {
		for key, val := range def[0] {
			// 如果 result 中不存在 key，则合并
			if _, ok := result[key]; !ok {
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

// 获取 *gin.Context.Get() 中的值
func (this base) get(ctx *gin.Context, key any, def ...any) (value any) {
	if item, exist := ctx.Get(cast.ToString(key)); exist {
		value = item
	} else {
		if empty := utils.Is.Empty(def); !empty {
			value = def[0]
		}
	}
	return
}

// ============================== cache ==============================

type cache struct{}

// API请求结果是否优先从缓存中获取
func (this cache) enable(ctx *gin.Context) (ok bool) {
	item := cast.ToBool(base{}.param(ctx, "cache", "true"))
	where := item && cast.ToBool(facade.CacheToml.Get("open"))
	return utils.Ternary[bool](where, true, false)
}

// 缓存名称
func (this cache) name(ctx *gin.Context) (name string) {

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

	// 生产缓存名称 - 禁止包含（兼容 Windows）：\/:*?"<>|
	path := strings.Replace(ctx.Request.URL.Path, "/", "_", -1)
	name = fmt.Sprintf("[%s]%s&hash=%s", ctx.Request.Method, path, facade.Hash.Sum32(name))
	return
}

// ============================== 上下文挂载的 meta 信息 ==============================

type meta struct{}

// 从上下文中解析用户信息
func (this meta) user(ctx *gin.Context) (result model.Users) {

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

// 分页限制
func (this meta) limit(ctx *gin.Context) (result int) {

	// 请求参数
	params := base{}.params(ctx, map[string]any{
		"limit": 10,
	})

	// 配置信息
	var config map[string]any

	// 缓存名称
	cacheName := "config[SYSTEM_PAGE_LIMIT]"
	// 是否开启了缓存
	cacheState := cast.ToBool(facade.CacheToml.Get("open"))

	// 检查缓存是否存在
	if cacheState && facade.Cache.Has(cacheName) {

		config = cast.ToStringMap(facade.Cache.Get(cacheName))

	} else {

		config = map[string]any{
			"value": 0,
			"text":  30,
		}
		// 存储到缓存中
		if cacheState {
			go facade.Cache.Set(cacheName, config)
		}
	}

	// 最大限制
	max := cast.ToInt(config["text"])
	// 当前限制
	limit := cast.ToInt(params["limit"])
	// 是否开启了限制
	state := cast.ToBool(config["value"])

	// 限制小于等于 0 - 返回默认值
	if limit <= 0 {
		return 10
	}

	// 没开启限制 - 直接返回
	if !state {
		return limit
	}

	if limit > max {
		return max
	}

	return limit
}

// ============================== 上下文挂载的 meta 信息 ==============================
