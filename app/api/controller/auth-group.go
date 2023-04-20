package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"inis/app/model"
	"inis/app/validator"
	"math"
	"strings"
	"time"
)

type AuthGroup struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *AuthGroup) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"one":    this.one,
		"all":    this.all,
		"count":  this.count,
		"column": this.column,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *AuthGroup) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"save":   this.save,
		"create": this.create,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *AuthGroup) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"update":  this.update,
		"restore": this.restore,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *AuthGroup) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"remove": this.remove,
		"delete": this.delete,
		"clear":  this.clear,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *AuthGroup) INDEX(ctx *gin.Context) {
	this.json(ctx, nil, facade.Lang(ctx, "没什么用！"), 202)
}

// one 获取指定数据
func (this *AuthGroup) one(ctx *gin.Context) {

	code := 204
	msg := []string{"无数据！", ""}
	var data any

	// 获取请求参数
	params := this.params(ctx, map[string]any{
		"order":       "create_time desc",
		"onlyTrashed": false,
		"withTrashed": false,
	})

	// 表数据结构体
	table := model.AuthGroup{}
	// 允许查询的字段
	allow := []any{"id"}
	// 动态给结构体赋值
	for key, val := range params {
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			utils.Struct.Set(&table, key, val)
		}
	}

	cacheName := this.cache.name(ctx)
	// 开启了缓存 并且 缓存中有数据
	if this.cache.enable(ctx) && facade.Cache.Has(cacheName) {

		// 从缓存中获取数据
		msg[1] = "（来自缓存）"
		data = facade.Cache.Get(cacheName)

	} else {

		mold := facade.DB.Model(&table).OnlyTrashed(params["onlyTrashed"]).WithTrashed(params["withTrashed"])
		mold.IWhere(params["where"]).IOr(params["or"]).ILike(params["like"]).INot(params["not"]).INull(params["null"]).INotNull(params["notNull"])
		item := mold.Where(table).Order(params["order"]).Find()
		// 缓存数据
		if this.cache.enable(ctx) {
			go facade.Cache.Set(cacheName, item)
		}
		data = item
	}

	if !utils.Is.Empty(data) {
		code = 200
		msg[0] = "数据请求成功！"
	}

	this.json(ctx, data, facade.Lang(ctx, strings.Join(msg, "")), code)
}

// all 获取全部数据
func (this *AuthGroup) all(ctx *gin.Context) {

	code := 204
	msg := []string{"无数据！", ""}
	var data any

	// 获取请求参数
	params := this.params(ctx, map[string]any{
		"page":        1,
		"limit":       5,
		"order":       "create_time desc",
		"onlyTrashed": false,
		"withTrashed": false,
	})

	// 表数据结构体
	table := model.AuthGroup{}
	// 允许查询的字段
	var allow []any
	// 动态给结构体赋值
	for key, val := range params {
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			params[key] = val
			utils.Struct.Set(&table, key, val)
		}
	}

	page := cast.ToInt(params["page"])
	limit := cast.ToInt(params["limit"])
	var result []model.AuthGroup
	mold := facade.DB.Model(&result).OnlyTrashed(params["onlyTrashed"]).WithTrashed(params["withTrashed"])
	mold.IWhere(params["where"]).IOr(params["or"]).ILike(params["like"]).INot(params["not"]).INull(params["null"]).INotNull(params["notNull"])
	count := mold.Where(table).Count()

	cacheName := this.cache.name(ctx)
	// 开启了缓存 并且 缓存中有数据
	if this.cache.enable(ctx) && facade.Cache.Has(cacheName) {

		// 从缓存中获取数据
		msg[1] = "（来自缓存）"
		data = facade.Cache.Get(cacheName)

	} else {

		// 从数据库中获取数据
		item := mold.Where(table).Limit(limit).Page(page).Order(params["order"]).Select()
		// 缓存数据
		if this.cache.enable(ctx) {
			go facade.Cache.Set(cacheName, item)
		}
		data = item
	}

	if data != nil {
		code = 200
		msg[0] = "数据请求成功！"
	}

	this.json(ctx, gin.H{
		"data":  data,
		"count": count,
		"page":  math.Ceil(float64(count) / float64(limit)),
	}, facade.Lang(ctx, strings.Join(msg, "")), code)
}

// save 保存数据 - 包含创建和更新
func (this *AuthGroup) save(ctx *gin.Context) {

	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["id"]) {
		this.create(ctx)
	} else {
		this.update(ctx)
	}
}

// create 创建数据
func (this *AuthGroup) create(ctx *gin.Context) {

	// 获取请求参数
	params := this.params(ctx)
	// 验证器
	err := validator.NewValid("auth-group", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 表数据结构体
	table := model.AuthGroup{CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()}
	allow := []any{"name", "rules", "uids", "remark"}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, "无权限！", 400)
		return
	}

	// 增加允许的字段
	allow = append(allow)

	// 动态给结构体赋值
	for key, val := range params {
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			utils.Struct.Set(&table, key, val)
		}
	}

	// 创建数据
	facade.DB.Model(&table).Create(&table)

	this.json(ctx, map[string]any{
		"id": table.Id,
	}, "创建成功！", 200)
}

// update 更新数据
func (this *AuthGroup) update(ctx *gin.Context) {

	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["id"]) {
		this.json(ctx, nil, "id不能为空！", 400)
		return
	}

	// 验证器
	err := validator.NewValid("auth-group", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 表数据结构体
	table := model.AuthGroup{}
	allow := []any{"name", "rules", "uids", "remark"}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, "无权限！", 400)
		return
	}

	// 增加允许的字段
	allow = append(allow)
	async := utils.Async[map[string]any]()

	// 动态给结构体赋值
	for key, val := range params {
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			async.Set(key, val)
		}
	}

	// 更新数据
	facade.DB.Model(&table).Force().Where("id", params["id"]).Update(async.Result())

	this.json(ctx, map[string]any{
		"id": table.Id,
	}, "更新成功！", 200)
}

// count 统计数据
func (this *AuthGroup) count(ctx *gin.Context) {

	// 表数据结构体
	table := model.AuthGroup{}
	// 获取请求参数
	params := this.params(ctx)

	item := facade.DB.Model(&table).OnlyTrashed(params["onlyTrashed"]).WithTrashed(params["withTrashed"])
	item.IWhere(params["where"]).IOr(params["or"]).ILike(params["like"]).INot(params["not"]).INull(params["null"]).INotNull(params["notNull"])

	count := item.Count()

	this.json(ctx, count, "查询成功！", 200)
}

// column 获取单列数据
func (this *AuthGroup) column(ctx *gin.Context) {

	// 表数据结构体
	table := model.AuthGroup{}
	// 获取请求参数
	params := this.params(ctx, map[string]any{
		"field": "*",
	})

	item := facade.DB.Model(&table).OnlyTrashed(params["onlyTrashed"]).WithTrashed(params["withTrashed"]).Order(params["order"])
	item.IWhere(params["where"]).IOr(params["or"]).ILike(params["like"]).INot(params["not"]).INull(params["null"]).INotNull(params["notNull"])

	if !strings.Contains(cast.ToString(params["field"]), "*") {
		item.Field(params["field"])
	}

	this.json(ctx, item.Column(), "查询成功！", 200)
}

// remove 软删除
func (this *AuthGroup) remove(ctx *gin.Context) {

	// 表数据结构体
	table := model.AuthGroup{}
	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["ids"]) {
		this.json(ctx, nil, "ids不能为空！", 400)
		return
	}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, "无权限！", 400)
		return
	}

	// 软删除
	facade.DB.Model(&table).Delete(params["ids"])

	this.json(ctx, nil, "删除成功！", 200)
}

// delete 真实删除
func (this *AuthGroup) delete(ctx *gin.Context) {

	// 表数据结构体
	table := model.AuthGroup{}
	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["ids"]) {
		this.json(ctx, nil, "ids不能为空！", 400)
		return
	}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, "无权限！", 400)
		return
	}

	// 真实删除
	facade.DB.Model(&table).Force().Delete(params["ids"])

	this.json(ctx, nil, "删除成功！", 200)
}

// clear 清空回收站
func (this *AuthGroup) clear(ctx *gin.Context) {

	// 表数据结构体
	table := model.AuthGroup{}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, "无权限！", 400)
		return
	}

	// 找到所有软删除的数据
	facade.DB.Model(&table).OnlyTrashed().Force().Delete()

	this.json(ctx, nil, "清空成功！", 200)
}

// restore 恢复数据
func (this *AuthGroup) restore(ctx *gin.Context) {

	// 表数据结构体
	table := model.AuthGroup{}
	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["ids"]) {
		this.json(ctx, nil, "ids不能为空！", 400)
		return
	}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, "无权限！", 400)
		return
	}

	// 还原数据
	facade.DB.Model(&table).Restore(params["ids"])

	this.json(ctx, nil, "恢复成功！", 200)
}
