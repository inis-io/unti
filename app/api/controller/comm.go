package controller

import (
	"fmt"
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
		"login":         this.login,
		"register":      this.register,
		"social-login":  this.socialLogin,
		"check-token":   this.checkToken,
		"reset-passowd": this.resetPassword,
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

	allow := map[string]any{
		"logout": this.logout,
	}
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

	if utils.Is.Empty(params["account"]) {
		this.json(ctx, nil, facade.Lang(ctx, "请提交帐号（或邮箱和手机号）！"), 400)
		return
	}

	if utils.Is.Empty(params["password"]) {
		this.json(ctx, nil, facade.Lang(ctx, "请提交密码！"), 400)
		return
	}

	// 正则表达式，匹配通过空格分割的两个16位任意字符 `^(\w{16}) (\w{16})$`
	reg := regexp.MustCompile(`^([\w+]{16})\D+([\w+]{16})$`)
	match := reg.FindStringSubmatch(ctx.GetHeader("i-cipher"))

	// 密文解密
	if match != nil {

		cipher := facade.Cipher(match[1], match[2])

		deAccount := cipher.Decrypt([]byte(cast.ToString(params["account"])))
		dePassword := cipher.Decrypt(params["password"])
		if deAccount.Error != nil || dePassword.Error != nil {
			this.json(ctx, nil, facade.Lang(ctx, "帐号或密码解密失败！"), 400)
			return
		}

		params["account"] = deAccount.Text
		params["password"] = dePassword.Text
	}

	// 查询用户是否存在
	item := facade.DB.Model(&table).Or([]any{
		[]any{"email", "=", params["account"]},
		[]any{"phone", "=", params["account"]},
		[]any{"account", "=", params["account"]},
	}).Where("source", params["source"]).Find()

	if utils.Is.Empty(item) {
		this.json(ctx, nil, facade.Lang(ctx, "账户不存在！"), 400)
		return
	}

	if utils.Is.Empty(table.Password) {
		this.json(ctx, nil, facade.Lang(ctx, "该帐号未设置密码，请切换登录方式！"), 400)
		return
	}

	// 密码校验
	if utils.Password.Verify(table.Password, params["password"]) == false {
		this.json(ctx, nil, facade.Lang(ctx, "密码错误！"), 400)
		return
	}

	jwt := facade.Jwt().Create(facade.H{
		"uid":  table.Id,
		"hash": facade.Hash.Sum32(table.Password),
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
		"token": jwt.Text,
	}

	// 往客户端写入cookie - 存储登录token
	setToken(ctx, jwt.Text)

	this.json(ctx, result, facade.Lang(ctx, "登录成功！"), 200)
}

// 注册
func (this *Comm) register(ctx *gin.Context) {

	if !cast.ToBool(this.signInConfig()["value"]) {
		this.json(ctx, nil, "管理员关闭了注册功能！", 403)
		return
	}

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

	if utils.Is.Empty(params["social"]) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 不能为空！", "social"), 400)
		return
	}

	var social string
	social = utils.Ternary(utils.Is.Email(params["social"]), "email", social)
	social = utils.Ternary(utils.Is.Phone(params["social"]), "phone", social)

	if utils.Is.Empty(social) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 格式不正确！", "social"), 400)
		return
	}

	// 判断是否已经注册
	ok := facade.DB.Model(&table).Where([]any{
		[]any{"source", "=", params["source"]},
		[]any{social, "=", params["social"]},
	}).Exist()
	// 已注册
	if ok {
		switch social {
		case "email":
			this.json(ctx, nil, facade.Lang(ctx, "该邮箱已经注册！"), 400)
			return
		case "phone":
			this.json(ctx, nil, facade.Lang(ctx, "该手机号已经注册！"), 400)
			return
		}
	}

	if !utils.Is.Empty(params["account"]) {
		// 判断账号是否已经注册
		ok := facade.DB.Model(&table).Where([]any{
			[]any{"source", "=", params["source"]},
			[]any{"account", "=", params["account"]},
		}).Exist()
		if ok {
			this.json(ctx, nil, facade.Lang(ctx, "该帐号已经注册！"), 400)
			return
		}
	}

	cacheName := fmt.Sprintf("%v-%v", social, params["social"])

	// 验证码为空 - 发送验证码
	if utils.Is.Empty(params["code"]) {

		drive := utils.Ternary(social == "email", "email", "sms")
		sms   := facade.NewSMS(drive).VerifyCode(params["social"])
		if sms.Error != nil {
			this.json(ctx, nil, sms.Error.Error(), 400)
			return
		}
		// 缓存验证码 - 5分钟
		facade.Cache.Set(cacheName, sms.VerifyCode, 5*time.Minute)
		this.json(ctx, nil, facade.Lang(ctx, "验证码发送成功！"), 201)
		return
	}

	if utils.Is.Empty(params["password"]) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 不能为空！", "密码"), 400)
		return
	}

	// 获取缓存里面的验证码
	cacheCode := facade.Cache.Get(cacheName)

	if cast.ToString(params["code"]) != cacheCode {
		this.json(ctx, nil, facade.Lang(ctx, "验证码错误！"), 400)
		return
	}

	// 允许存储的字段
	allow := []any{"account", "password", "email", "phone", "nickname", "avatar", "description", "source"}
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
	utils.Struct.Set(&table, social, params["social"])

	// 设置登录时间
	utils.Struct.Set(&table, "login_time", time.Now().Unix())

	// 创建用户
	tx := facade.DB.Model(&table).Create(&table)
	if tx.Error != nil {
		this.json(ctx, nil, tx.Error.Error(), 400)
		return
	}

	// 删除验证码
	facade.Cache.Del(cacheName)

	jwt := facade.Jwt().Create(facade.H{
		"uid":  table.Id,
		"hash": facade.Hash.Sum32(table.Password),
	})

	// 删除密码
	table.Password = ""

	result := map[string]any{
		"user":  table,
		"token": jwt.Text,
	}

	// 往客户端写入cookie - 存储登录token
	setToken(ctx, jwt.Text)

	this.json(ctx, result, facade.Lang(ctx, "注册成功！"), 200)
}

// 社交方式登录 - 邮箱、手机号
func (this *Comm) socialLogin(ctx *gin.Context) {

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

	if utils.Is.Empty(params["social"]) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 不能为空！", "social"), 400)
		return
	}

	var social string
	social = utils.Ternary(utils.Is.Email(params["social"]), "email", social)
	social = utils.Ternary(utils.Is.Phone(params["social"]), "phone", social)

	if utils.Is.Empty(social) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 格式不正确！", "social"), 400)
		return
	}

	// 判断是否已经注册
	ok := facade.DB.Model(&table).Where([]any{
		[]any{"source", "=", params["source"]},
		[]any{social, "=", params["social"]},
	}).Exist()
	// 未注册 - 自动注册
	if !ok {

		if !cast.ToBool(this.signInConfig()["value"]) {
			this.json(ctx, nil, "请联系管理员为您手动开通账号！", 400)
			return
		}

		user := &model.Users{
			Account:  cast.ToString(params["email"]),
			Nickname: "会员" + utils.Rand.String(4, "0123456789"),
			Source:   cast.ToString(params["source"]),
		}

		switch social {
		case "email":
			user.Email = cast.ToString(params["social"])
		case "phone":
			user.Phone = cast.ToString(params["social"])
		}

		facade.DB.Model(&table).Create(user)
	}

	cacheName := fmt.Sprintf("%v-%v", social, params["social"])

	// 验证码为空 - 发送验证码
	if utils.Is.Empty(params["code"]) {

		drive := utils.Ternary(social == "email", "email", "sms")
		sms   := facade.NewSMS(drive).VerifyCode(params["social"])
		if sms.Error != nil {
			this.json(ctx, nil, sms.Error.Error(), 400)
			return
		}
		// 缓存验证码 - 5分钟
		facade.Cache.Set(cacheName, sms.VerifyCode, 5*time.Minute)
		this.json(ctx, nil, facade.Lang(ctx, "验证码发送成功！"), 201)
		return
	}

	// 获取缓存里面的验证码
	cacheCode := facade.Cache.Get(cacheName)

	if cast.ToString(params["code"]) != cacheCode {
		this.json(ctx, nil, facade.Lang(ctx, "验证码错误！"), 400)
		return
	}

	// 删除验证码
	facade.Cache.Del(cacheName)

	// 查询用户
	item := facade.DB.Model(&table).Where(social, params["social"]).Find()

	jwt := facade.Jwt().Create(facade.H{
		"uid":  table.Id,
		"hash": facade.Hash.Sum32(table.Password),
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
		"token": jwt.Text,
	}

	// 往客户端写入cookie - 存储登录token
	setToken(ctx, jwt.Text)

	this.json(ctx, result, facade.Lang(ctx, "登录成功！"), 200)
}

// 忘记密码
func (this *Comm) resetPassword(ctx *gin.Context) {

	// 表数据结构体
	table := model.Users{}
	// 请求参数
	params := this.params(ctx, map[string]any{
		"source": "default",
	})

	// 必须有一个不能为空
	if utils.Is.Empty(params["account"]) && utils.Is.Empty(params["social"]) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 和 %s 必须有一个不能为空！", "account", "social"), 400)
		return
	}

	// 验证器
	err := validator.NewValid("users", params)

	// 参数校验不通过
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 检查类型，邮箱或者手机号
	var social string
	if !utils.Is.Empty(params["social"]) {
		social = utils.Ternary(utils.Is.Email(params["social"]), "email", social)
		social = utils.Ternary(utils.Is.Phone(params["social"]), "phone", social)
	}

	var user map[string]any

	// 账号优先
	if !utils.Is.Empty(params["account"]) {

		// 判断账号是否已经注册
		user = facade.DB.Model(&table).Where("source", params["source"]).Where("account", params["account"]).Find()
		if utils.Is.Empty(user) {
			this.json(ctx, nil, facade.Lang(ctx, "该账号未注册！"), 400)
			return
		}

		// 找回密码
		this.password(ctx, user)
		return
	}

	// 判断是否已经注册
	user = facade.DB.Model(&table).Where("source", params["source"]).Where(social, params["social"]).Find()
	if utils.Is.Empty(user) {
		switch social {
		case "email":
			this.json(ctx, nil, facade.Lang(ctx, "该邮箱未注册！"), 400)
			return
		case "phone":
			this.json(ctx, nil, facade.Lang(ctx, "该手机号未注册！"), 400)
			return
		}
	}

	// 找回密码
	this.password(ctx, user)
}

// 忘记密码
func (this *Comm) password(ctx *gin.Context, user map[string]any) {

	// 请求参数
	params := this.params(ctx)

	drives := cast.ToStringMap(facade.SMSToml.Get("drive"))

	// 驱动、社交、驱动模式
	var drive, social, mode string

	// 邮箱驱动 - 次之
	if !utils.Is.Empty(drives["email"]) && !utils.Is.Empty(user["email"]) {
		mode   = "email"
		drive  = cast.ToString(drives["email"])
		social = cast.ToString(user["email"])
	}

	// SMS驱动 - 优先 - 覆盖
	if !utils.Is.Empty(drives["sms"]) && !utils.Is.Empty(user["phone"]) {
		mode   = "sms"
		drive  = cast.ToString(drives["sms"])
		social = cast.ToString(user["phone"])
	}

	// 如果提交了 social
	if !utils.Is.Empty(params["social"]) {
		var unknown string
		unknown = utils.Ternary(utils.Is.Email(params["social"]), "email", mode)
		unknown = utils.Ternary(utils.Is.Phone(params["social"]), "sms", mode)
		// 如果驱动存在，且提交的 social 也存在
		if !utils.Is.Empty(drives[mode]) && !utils.Is.Empty(unknown) {
			mode   = unknown
			social = cast.ToString(params["social"])
			drive  = cast.ToString(drives[mode])
		}
	}

	// 都不满足
	if utils.Is.Empty(drive) {

		// 既没开启邮箱驱动，也没开启SMS驱动
		if utils.Is.Empty(drives["email"]) && utils.Is.Empty(drives["sms"]) {
			this.json(ctx, nil, facade.Lang(ctx, "请联系管理员重置密码！"), 400)
			return
		}

		if !utils.Is.Empty(user["phone"]) {
			this.json(ctx, nil, facade.Lang(ctx, "管理员未开启短信服务，无法发送验证码！"), 400)
			return
		}

		if !utils.Is.Empty(user["email"]) {
			this.json(ctx, nil, facade.Lang(ctx, "管理员未开启邮箱服务，无法发送验证码！"), 400)
			return
		}

		this.json(ctx, nil, facade.Lang(ctx, "请联系管理员重置密码！"), 400)
		return
	}

	// 缓存名称
	cacheName := fmt.Sprintf("%v-%v", drive, social)

	// 验证码为空 - 发送验证码
	if utils.Is.Empty(params["code"]) {

		sms := facade.NewSMS(drive).VerifyCode(social)
		if sms.Error != nil {
			this.json(ctx, nil, sms.Error.Error(), 400)
			return
		}
		// 缓存验证码 - 5分钟
		go facade.Cache.Set(cacheName, sms.VerifyCode, 5 * time.Minute)

		msg := fmt.Sprintf("验证码发送至您的%v：%s，请注意查收！", utils.Ternary(mode == "email", "邮箱", "手机"), social)
		this.json(ctx, nil, facade.Lang(ctx, msg), 201)
		return
	}

	if utils.Is.Empty(params["password"]) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 不能为空！", "密码"), 400)
		return
	}

	// 获取缓存里面的验证码
	cacheCode := facade.Cache.Get(cacheName)

	if cast.ToString(params["code"]) != cast.ToString(cacheCode) {
		this.json(ctx, nil, facade.Lang(ctx, "验证码错误！"), 400)
		return
	}

	// 加密密码
	password := utils.Password.Create(params["password"])

	// 更新密码
	tx := facade.DB.Model(&model.Users{}).Where("id", user["id"]).UpdateColumn("password", password)
	if tx.Error != nil {
		this.json(ctx, nil, tx.Error.Error(), 400)
		return
	}

	// 删除验证码
	go facade.Cache.Del(cacheName)

	this.json(ctx, nil, facade.Lang(ctx, "密码重置成功！"), 200)
}

// 校验token
func (this *Comm) checkToken(ctx *gin.Context) {

	params := this.params(ctx)

	tokenName := cast.ToString(facade.AppToml.Get("app.token_name", "INIS_LOGIN_TOKEN"))

	var token string
	if !utils.Is.Empty(ctx.Request.Header.Get("Authorization")) {
		token = ctx.Request.Header.Get("Authorization")
	} else {
		token, _ = ctx.Cookie(tokenName)
	}

	if utils.Is.Empty(token) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 不能为空！", "Authorization"), 412)
		return
	}

	// 解析token
	jwt := facade.Jwt().Parse(token)
	if jwt.Error != nil {
		this.json(ctx, nil, facade.Lang(ctx, "%s 无效！", "Authorization"), 400)
		return
	}

	// 表数据结构体
	table := model.Users{}
	// 查询用户
	item := facade.DB.Model(&table).Where("id", jwt.Data["uid"]).Find()
	if utils.Is.Empty(item) {
		this.json(ctx, nil, facade.Lang(ctx, "用户不存在！"), 204)
		return
	}

	// token 有效时长
	valid := jwt.Valid

	if cast.ToBool(params["renew"]) {
		jwt = facade.Jwt().Create(facade.H{
			"uid":  table.Id,
			"hash": facade.Hash.Sum32(table.Password),
		})
		valid = cast.ToInt64(utils.Calc(facade.AppToml.Get("jwt.expire", "7200")))
		// 往客户端写入cookie - 存储登录token
		setToken(ctx, jwt.Text)
	}

	delete(item, "password")

	this.json(ctx, gin.H{
		"user":       item,
		"token":      jwt.Text,
		"valid_time": valid,
	}, facade.Lang(ctx, facade.Lang(ctx, "合法的token！")), 200)
}

// 退出登录
func (this *Comm) logout(ctx *gin.Context) {
	ctx.SetCookie(cast.ToString(facade.AppToml.Get("app.token_name", "INIS_LOGIN_TOKEN")), "", -1, "/", "", false, false)
	this.json(ctx, nil, facade.Lang(ctx, "退出成功！"), 200)
}

// 设置登录token到客户的cookie中
func setToken(ctx *gin.Context, token any) {

	host := ctx.Request.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	expire := cast.ToInt(facade.AppToml.Get("jwt.expire", "7200"))
	tokenName := cast.ToString(facade.AppToml.Get("app.token_name", "INIS_LOGIN_TOKEN"))

	ctx.SetCookie(tokenName, cast.ToString(token), expire, "/", host, false, false)
}

// 获取注册配置
func (this *Comm) signInConfig() (result map[string]any) {
	return map[string]any{
		"value": 1,
	}
}
