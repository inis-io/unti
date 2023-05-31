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

type Users struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *Users) IGET(ctx *gin.Context) {
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
func (this *Users) IPOST(ctx *gin.Context) {

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
func (this *Users) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"update" : this.update,
		"restore": this.restore,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *Users) IDEL(ctx *gin.Context) {
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
func (this *Users) INDEX(ctx *gin.Context) {
	this.json(ctx, nil, facade.Lang(ctx, "没什么用！"), 202)
}

// one 获取指定数据
func (this *Users) one(ctx *gin.Context) {

	code := 204
	msg  := []string{"无数据！", ""}
	var data any

	// 获取请求参数
	params := this.params(ctx, map[string]any{
		"order":       "create_time desc",
		"onlyTrashed": false,
		"withTrashed": false,
	})

	// 表数据结构体
	table := model.Users{}
	// 允许查询的字段
	allow := []any{"id", "account", "email"}
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
		// 删除指定字段
		delete(item, "password")
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
func (this *Users) all(ctx *gin.Context) {

	code := 204
	msg  := []string{"无数据！", ""}
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
	table := model.Users{}
	// 允许查询的字段
	allow := []any{"level", "source"}
	// 动态给结构体赋值
	for key, val := range params {
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			params[key] = val
			utils.Struct.Set(&table, key, val)
		}
	}

	page  := cast.ToInt(params["page"])
	limit := cast.ToInt(params["limit"])
	var result []model.Users
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
		// 删除指定字段
		for _, val := range item {
			delete(val, "password")
		}
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
func (this *Users) save(ctx *gin.Context) {

	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["id"]) {
		this.create(ctx)
	} else {
		this.update(ctx)
	}
}

// create 创建数据
func (this *Users) create(ctx *gin.Context) {

	// 获取请求参数
	params := this.params(ctx)
	// 验证器
	err := validator.NewValid("users", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 表数据结构体
	table := model.Users{CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()}
	allow := []any{"account", "password", "nickname", "email", "avatar", "description"}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	if utils.Is.Empty(params["email"]) {
		this.json(ctx, nil, facade.Lang(ctx, "邮箱不能为空！"), 400)
		return
	}

	// 判断邮箱是否已经注册
	ok := facade.DB.Model(&table).Where("email", params["email"]).FindOrEmpty()
	if !ok {
		this.json(ctx, nil, facade.Lang(ctx, "该邮箱已经注册！"), 400)
		return
	}

	// 判断账号是否已经注册
	if !utils.Is.Empty(params["account"]) {
		ok := facade.DB.Model(&table).Where("account", params["account"]).FindOrEmpty()
		if !ok {
			this.json(ctx, nil, facade.Lang(ctx, "该账号已经注册！"), 400)
			return
		}
	}

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, facade.Lang(ctx, "无权限！"), 401)
		return
	}

	// 增加允许的字段
	allow = append(allow, "level", "source")

	// 动态给结构体赋值
	for key, val := range params {
		// 加密密码
		if key == "password" {
			val = utils.Password.Create(params["password"])
		}
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			utils.Struct.Set(&table, key, val)
		}
	}

	// 创建用户
	facade.DB.Model(&table).Create(&table)

	this.json(ctx, map[string]any{
		"id": table.Id,
	}, facade.Lang(ctx, "创建成功！"), 200)
}

// update 更新数据
func (this *Users) update(ctx *gin.Context) {

	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["id"]) {
		this.json(ctx, nil, facade.Lang(ctx, "id不能为空！"), 400)
		return
	}

	// 验证器
	err := validator.NewValid("users", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 表数据结构体
	table := model.Users{}
	allow := []any{"id", "account", "password", "nickname", "email", "avatar", "description"}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {

		if cast.ToInt(user["id"]) != cast.ToInt(params["id"]) {
			this.json(ctx, nil, facade.Lang(ctx, "无权限！"), 401)
			return
		}
	}

	// 增加允许的字段
	allow = append(allow, "level", "source")
	async := utils.Async[map[string]any]()

	// 动态给结构体赋值
	for key, val := range params {
		// 加密密码
		if key == "password" {
			val = utils.Password.Create(params["password"])
		}
		// 防止恶意传入字段
		if utils.In.Array(key, allow) {
			async.Set(key, val)
		}
	}

	// 账号唯一处理
	if !utils.Is.Empty(params["account"]) {
		item := facade.DB.Model(&table).Where("account", params["account"]).Find()
		if item != nil && cast.ToInt(item["id"]) != cast.ToInt(params["id"]) {
			this.json(ctx, nil, facade.Lang(ctx, "帐号已存在！"), 400)
			return
		}
	}

	// 邮箱唯一处理
	if !utils.Is.Empty(params["email"]) {
		ok := facade.DB.Model(&table).Where([]any{
			[]any{"id", "<>", params["id"]},
			[]any{"email", "=", params["email"]},
		}).FindOrEmpty()
		if !ok {
			this.json(ctx, nil, facade.Lang(ctx, "邮箱已存在！"), 400)
			return
		}
	}

	// 更新用户
	facade.DB.Model(&table).Force().Where("id", params["id"]).Update(async.Result())

	this.json(ctx, map[string]any{
		"id": table.Id,
	}, facade.Lang(ctx, "更新成功！"), 200)
}

// count 统计数据
func (this *Users) count(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 获取请求参数
	params := this.params(ctx)

	item := facade.DB.Model(&table).OnlyTrashed(params["onlyTrashed"]).WithTrashed(params["withTrashed"])
	item.IWhere(params["where"]).IOr(params["or"]).ILike(params["like"]).INot(params["not"]).INull(params["null"]).INotNull(params["notNull"])

	count := item.Count()

	this.json(ctx, count, facade.Lang(ctx, "查询成功！"), 200)
}

// column 获取单列数据
func (this *Users) column(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 获取请求参数
	params := this.params(ctx, map[string]any{
		"field": "*",
	})

	item := facade.DB.Model(&table).OnlyTrashed(params["onlyTrashed"]).WithTrashed(params["withTrashed"]).Order(params["order"])
	item.IWhere(params["where"]).IOr(params["or"]).ILike(params["like"]).INot(params["not"]).INull(params["null"]).INotNull(params["notNull"])

	// 排除密码
	if strings.Contains(cast.ToString(params["field"]), "password") {
		// 转换为数组
		field := strings.Split(cast.ToString(params["field"]), ",")
		// 排除密码
		field = utils.Array.Remove(field, "password")
		// 转换为字符串
		params["field"] = strings.Join(field, ",")
	}

	if !strings.Contains(cast.ToString(params["field"]), "*") {
		item.Field(params["field"])
	}

	item.WithoutField("password")

	this.json(ctx, item.Column(), facade.Lang(ctx, "查询成功！"), 200)
}

// remove 软删除
func (this *Users) remove(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["ids"]) {
		this.json(ctx, nil, facade.Lang(ctx, "ids不能为空！"), 400)
		return
	}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {

		this.json(ctx, nil, facade.Lang(ctx, "无权限！"), 401)
		return
	}

	// 软删除
	facade.DB.Model(&table).Delete(params["ids"])

	this.json(ctx, nil, facade.Lang(ctx, "删除成功！"), 200)
}

// delete 真实删除
func (this *Users) delete(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["ids"]) {
		this.json(ctx, nil, facade.Lang(ctx, "ids不能为空！"), 400)
		return
	}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {

		this.json(ctx, nil, facade.Lang(ctx, "无权限！"), 401)
		return
	}

	// 真实删除
	facade.DB.Model(&table).Force().Delete(params["ids"])

	this.json(ctx, nil, facade.Lang(ctx, "删除成功！"), 200)
}

// clear 清空回收站
func (this *Users) clear(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, facade.Lang(ctx, "无权限！"), 401)
		return
	}

	// 找到所有软删除的数据
	facade.DB.Model(&table).OnlyTrashed().Force().Delete()

	this.json(ctx, nil, facade.Lang(ctx, "清空成功！"), 200)
}

// restore 恢复数据
func (this *Users) restore(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 获取请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["ids"]) {
		this.json(ctx, nil, facade.Lang(ctx, "ids不能为空！"), 400)
		return
	}

	token, _ := ctx.Get("user")
	user := cast.ToStringMap(token)

	// 权限判断
	if !utils.In.Array(user["level"], []any{"admin"}) {
		this.json(ctx, nil, facade.Lang(ctx, "无权限！"), 401)
		return
	}

	// 还原数据
	facade.DB.Model(&table).Restore(params["ids"])

	this.json(ctx, nil, facade.Lang(ctx, "恢复成功！"), 200)
}
