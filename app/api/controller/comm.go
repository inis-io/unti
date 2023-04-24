package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"inis/app/model"
	"inis/app/validator"
	"regexp"
	"strings"
	"time"
)

type Comm struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *Comm) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *Comm) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"login":       this.login,
		"register":    this.register,
		"email-login": this.emailLogin,
		"check-token": this.checkToken,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *Comm) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *Comm) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *Comm) INDEX(ctx *gin.Context) {
	this.json(ctx, nil, facade.Lang(ctx, "没什么用！"), 202)
}

// 登录
func (this *Comm) login(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 请求参数
	params := this.params(ctx, map[string]any{
		"source": "default",
	})

	// 验证器
	err := validator.NewValid("users", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	if utils.Is.Empty(params["account"]) {
		this.json(ctx, nil, "请提交帐号（或邮箱）！", 400)
		return
	}

	if utils.Is.Empty(params["password"]) {
		this.json(ctx, nil, "请提交密码！", 400)
		return
	}

	// 正则表达式，匹配通过空格分割的两个16位任意字符 `^(\w{16}) (\w{16})$`
	reg   := regexp.MustCompile(`^([\w+]{16})\D+([\w+]{16})$`)
	match := reg.FindStringSubmatch(ctx.GetHeader("i-cipher"))

	// 密文解密
	if match != nil {

		cipher := facade.Cipher(match[1], match[2])

		deAccount  := cipher.Decrypt([]byte(cast.ToString(params["account"])))
		dePassword := cipher.Decrypt(params["password"])
		if deAccount.Error != nil || dePassword.Error != nil {
			this.json(ctx, nil, "帐号或密码解密失败！", 400)
			return
		}

		params["account"]  = deAccount.Text
		params["password"] = dePassword.Text
	}

	// 查询用户是否存在
	item := facade.DB.Model(&table).Or([]any{
		[]any{"email", "=", params["account"]},
		[]any{"account", "=", params["account"]},
	}).Where("source", params["source"]).Find()

	if item == nil {
		this.json(ctx, nil, "账户信息不存在！", 400)
		return
	}

	if utils.Is.Empty(table.Password) {
		this.json(ctx, nil, "该帐号未设置密码，请切换登录方式！", 400)
		return
	}

	// 密码校验
	if utils.Password.Verify(table.Password, params["password"]) == false {
		this.json(ctx, nil, "密码错误！", 400)
		return
	}

	token, _ := facade.Jwt.Create(map[string]any{
		"uid": table.Id,
	})

	// 删除 item 中的密码
	delete(item, "password")
	// 更新用户登录时间
	item["login_time"] = time.Now().Unix()
	facade.DB.Model(&table).Where("id", table.Id).Update(map[string]any{
		"login_time": item["login_time"],
	})

	result := map[string]any{
		"user":  item,
	}

	// 往客户端写入cookie - 存储登录token
	setToken(ctx, token)

	this.json(ctx, result, "登录成功！", 200)
}

// 注册
func (this *Comm) register(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 请求参数
	params := this.params(ctx, map[string]any{
		"source": "default",
	})

	// 验证器
	err := validator.NewValid("users", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	if utils.Is.Empty(params["email"]) {
		this.json(ctx, nil, "邮箱不能为空！", 400)
		return
	}

	// 判断邮箱是否已经注册
	ok := facade.DB.Model(&table).Where([]any{
		[]any{"source", "=", params["source"]},
		[]any{"email", "=", params["email"]},
	}).FindOrEmpty()
	if ok {
		this.json(ctx, nil, "该邮箱已经注册！", 400)
		return
	}

	if !utils.Is.Empty(params["account"]) {
		// 判断帐号是否已经注册
		ok := facade.DB.Model(&table).Where([]any{
			[]any{"source", "=", params["source"]},
			[]any{"account", "=", params["account"]},
		}).FindOrEmpty()
		if ok {
			this.json(ctx, nil, "该帐号已经注册！", 400)
			return
		}
	}

	email     := cast.ToString(params["email"])
	cacheName := "email-" + email

	// 验证码为空 - 发送验证码
	if utils.Is.Empty(params["code"]) {

		sms := facade.SMS.VerifyCode(email)
		if sms.Error != nil {
			this.json(ctx, nil, sms.Error.Error(), 400)
			return
		}
		// 缓存验证码 - 5分钟
		facade.Cache.Set(cacheName, sms.VerifyCode, 5 * time.Minute)
		this.json(ctx, nil, "验证码发送成功！", 201)
		return
	}

	if utils.Is.Empty(params["password"]) {
		this.json(ctx, nil, "密码不能为空！", 400)
		return
	}

	// 获取缓存里面的验证码
	cacheCode := facade.Cache.Get(cacheName)

	if cast.ToString(params["code"]) != cacheCode {
		this.json(ctx, nil, "验证码错误！", 400)
		return
	}

	// 允许存储的字段
	allow := []any{"account", "password", "email", "nickname", "avatar", "description", "source"}
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

	// 设置登录时间
	utils.Struct.Set(&table, "login_time", time.Now().Unix())

	// 创建用户
	facade.DB.Model(&table).Create(&table)

	// 删除验证码
	facade.Cache.Del(cacheName)

	token, _ := facade.Jwt.Create(map[string]any{
		"uid": table.Id,
	})

	// 删除密码
	table.Password = ""

	result := map[string]any{
		"user":  table,
	}

	// 往客户端写入cookie - 存储登录token
	setToken(ctx, token)

	this.json(ctx, result, "注册成功！", 200)
}

// 邮箱登录
func (this *Comm) emailLogin(ctx *gin.Context) {

	table := model.Users{}
	params := this.params(ctx, map[string]any{
		"source": "default",
	})

	// 验证器
	err := validator.NewValid("users", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// social - contact
	if utils.Is.Empty(params["email"]) {
		this.json(ctx, nil, "邮箱不能为空！", 400)
		return
	}

	// 判断邮箱是否已经注册
	ok := facade.DB.Model(&table).Where([]any{
		[]any{"source", "=", params["source"]},
		[]any{"email", "=", params["email"]},
	}).FindOrEmpty()
	if !ok {
		facade.DB.Model(&table).Create(&model.Users{
			Account:  cast.ToString(params["email"]),
			Nickname: "会员" + utils.Rand.String(4, "0123456789"),
			Email:    cast.ToString(params["email"]),
			Source:   cast.ToString(params["source"]),
		})
	}

	email := cast.ToString(params["email"])
	cacheName := "email-" + email

	// 验证码为空 - 发送验证码
	if utils.Is.Empty(params["code"]) {

		sms := facade.SMS.VerifyCode(email)
		if sms.Error != nil {
			this.json(ctx, nil, sms.Error.Error(), 400)
			return
		}
		// 缓存验证码 - 5分钟
		facade.Cache.Set(cacheName, sms.VerifyCode, 5 * time.Minute)
		this.json(ctx, nil, "验证码发送成功！", 201)
		return
	}

	// 获取缓存里面的验证码
	cacheCode := facade.Cache.Get(cacheName)

	if cast.ToString(params["code"]) != cacheCode {
		this.json(ctx, nil, "验证码错误！", 400)
		return
	}

	// 删除验证码
	facade.Cache.Del(cacheName)

	// 查询用户
	item := facade.DB.Model(&table).Where("email", params["email"]).Find()

	token, _ := facade.Jwt.Create(map[string]any{
		"uid": table.Id,
	})

	// 删除密码
	delete(item, "password")
	// 更新用户登录时间
	item["login_time"] = time.Now().Unix()
	facade.DB.Model(&table).Where("id", table.Id).Update(map[string]any{
		"login_time": item["login_time"],
	})

	result := map[string]any{
		"user":  item,
	}

	// 往客户端写入cookie - 存储登录token
	setToken(ctx, token)

	this.json(ctx, result, "登录成功！", 200)
}

// 校验token
func (this *Comm) checkToken(ctx *gin.Context) {

	var token string
	if !utils.Is.Empty(ctx.Request.Header.Get("Authorization")) {
		token = ctx.Request.Header.Get("Authorization")
	} else {
		token, _ = ctx.Cookie("UNTI_LOGIN_TOKEN")
	}

	if utils.Is.Empty(token) {
		this.json(ctx, nil, "Authorization不能为空！", 400)
		return
	}

	// 解析token
	jwt := facade.Jwt.Parse(token)
	if jwt.Error != nil {
		this.json(ctx, nil, "Authorization无效！", 400)
		return
	}

	// 表数据结构体
	table := model.Users{}
	// 查询用户
	item := facade.DB.Model(&table).Where("id", jwt.Data["uid"]).Find()
	if item == nil {
		this.json(ctx, nil, "用户不存在！", 204)
		return
	}

	delete(item, "password")

	this.json(ctx, map[string]any{
		"user": item,
	}, "合法的token！", 200)
}

// 设置登录token到客户的cookie中
func setToken(ctx *gin.Context, token any) {
	ctx.SetCookie(cast.ToString(facade.AppToml.Get("app.token_name", "UNTI_LOGIN_TOKEN")), cast.ToString(token), cast.ToInt(facade.AppToml.Get("jwt.expire", "7200")), "/", "", false, false)
}
