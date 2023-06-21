package controller

import (
	"context"
	"fmt"
	// JWT "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	JWT "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"mime/multipart"
	"strings"
	"time"
)

type Test struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *Test) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"request": this.request,
		"alipay":  this.alipay,
		"system":  this.system,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *Test) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"return-url": this.returnUrl,
		"notify-url": this.notifyUrl,
		"request":    this.request,
		"upload":     this.upload,
		"qq":         this.qq,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *Test) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"request": this.request,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *Test) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"request": this.request,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *Test) INDEX(ctx *gin.Context) {

	// 请求参数
	// params := this.params(ctx)

	res := gin.H{
		// "root" : this.meta.root(ctx),
		"user" : this.meta.user(ctx),
		// "route": this.meta.route(ctx),
		// "rules": this.meta.rules(ctx),
		// "json" : utils.Json.Encode(params["json"]),
	}

	this.json(ctx, res, facade.Lang(ctx, "好的！"), 200)
}

type JwtStruct struct {
	request  JwtRequest
	response JwtResponse
}

// JwtRequest - JWT请求
type JwtRequest struct {
	// 过期时间
	Expire  int64        `json:"expire"`
	// 颁发者签名
	Issuer  string       `json:"issuer"`
	// 主题
	Subject string       `json:"subject"`
	// 密钥
	Key     string       `json:"key"`
}

// JwtResponse - JWT响应
type JwtResponse struct {
	Text  string         `json:"text"`
	Data  map[string]any `json:"data"`
	Error error          `json:"error"`
	Valid int64          `json:"valid"`
}

// Jwt - 入口
func Jwt(request ...JwtRequest) *JwtStruct {

	if len(request) == 0 {
		request = append(request, JwtRequest{})
	}

	// 过期时间
	if request[0].Expire == 0 {
		request[0].Expire = cast.ToInt64(utils.Calc(facade.CryptToml.Get("jwt.expire", "7200")))
	}

	// 颁发者签名
	if utils.Is.Empty(request[0].Issuer) {
		request[0].Issuer = cast.ToString(facade.CryptToml.Get("jwt.issuer", "inis.cn"))
	}

	// 主题
	if utils.Is.Empty(request[0].Subject) {
		request[0].Subject = cast.ToString(facade.CryptToml.Get("jwt.subject", "inis"))
	}

	// 密钥
	if utils.Is.Empty(request[0].Key) {
		request[0].Key = cast.ToString(facade.CryptToml.Get("jwt.key", "inis"))
	}

	return &JwtStruct{
		request: request[0],
		response: JwtResponse{
			Data: make(map[string]any),
		},
	}
}

// Create - 创建JWT
func (this *JwtStruct) Create(data map[string]any) (result JwtResponse) {

	type JwtClaims struct {
		Data map[string]any `json:"data"`
		JWT.RegisteredClaims
	}

	IssuedAt  := JWT.NewNumericDate(time.Now())
	ExpiresAt := JWT.NewNumericDate(time.Now().Add(time.Second * time.Duration(this.request.Expire)))

	item, err := JWT.NewWithClaims(JWT.SigningMethodHS256, JwtClaims{
		Data: data,
		RegisteredClaims: JWT.RegisteredClaims{
			IssuedAt:  IssuedAt,				// 签发时间戳
			ExpiresAt: ExpiresAt,				// 过期时间戳
			Issuer:    this.request.Issuer,		// 颁发者签名
			Subject:   this.request.Subject,	// 签名主题
		},
	}).SignedString([]byte(this.request.Key))

	if err != nil {
		this.response.Error = err
		return this.response
	}

	this.response.Text = item

	return this.response
}

// Parse - 解析JWT
func (this *JwtStruct) Parse(token any) (result JwtResponse) {

	type JwtClaims struct {
		Data map[string]any `json:"data"`
		JWT.RegisteredClaims
	}

	item, err := JWT.ParseWithClaims(cast.ToString(token), &JwtClaims{}, func(token *JWT.Token) (any, error) {
		return []byte(this.request.Key), nil
	})

	if err != nil {
		facade.Log.Error(map[string]any{
			"error":     err,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "JWT解析错误")
		this.response.Error = err
		return this.response
	}

	if key, _ := item.Claims.(*JwtClaims); item.Valid {
		this.response.Data  = key.Data
		this.response.Valid = key.RegisteredClaims.ExpiresAt.Time.Unix() - time.Now().Unix()
	}

	return this.response
}

func (this *Test) qq(ctx *gin.Context) {

	params := this.params(ctx)

	if params["message_type"] == "private" {
		fmt.Println(utils.Json.Encode(params))

		item := utils.Curl(utils.CurlRequest{
			Method: "GET",
			Url:    "http://localhost:5700/send_private_msg",
			Query: map[string]any{
				"user_id": cast.ToString(params["user_id"]),
				"message": cast.ToString(params["message"]),
			},
		}).Send()

		if item.Error != nil {
			fmt.Println("发送失败", item.Error.Error())
			return
		}

		fmt.Println("发送成功", item.Json)
	}

	this.json(ctx, params, facade.Lang(ctx, "好的！"), 200)
}

func (this *Test) system(ctx *gin.Context) {

	params := this.params(ctx)

	this.json(ctx, params, facade.Lang(ctx, "好的！"), 200)
}

// INDEX - GET请求本体
func (this *Test) upload(ctx *gin.Context) {

	params := this.params(ctx)

	// 上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 文件数据
	bytes, err := file.Open()
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}
	defer func(bytes multipart.File) {
		err := bytes.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(bytes)

	// 文件后缀
	suffix := file.Filename[strings.LastIndex(file.Filename, "."):]
	params["suffix"] = suffix

	item := facade.Storage.Upload(facade.Storage.Path()+suffix, bytes)
	if item.Error != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	params["item"] = item

	fmt.Println("url: ", item.Domain+item.Path)

	this.json(ctx, params, facade.Lang(ctx, "好的！"), 200)
}

func (this *Test) alipay(ctx *gin.Context) {

	// 初始化 BodyMap
	body := make(gopay.BodyMap)
	body.Set("subject", "统一收单下单并支付页面接口")
	body.Set("out_trade_no", uuid.New().String())
	body.Set("total_amount", "0.01")
	body.Set("product_code", "FAST_INSTANT_TRADE_PAY")

	payUrl, err := facade.Alipay().TradePagePay(context.Background(), body)
	if err != nil {
		if bizErr, ok := alipay.IsBizError(err); ok {
			fmt.Println(bizErr)
			return
		}
		fmt.Println(err)
		return
	}

	fmt.Println(payUrl)

	this.json(ctx, payUrl, "数据请求成功！", 200)
}

func (this *Test) returnUrl(ctx *gin.Context) {

	params := this.params(ctx)

	fmt.Println("==================== returnUrl：", params)
}

func (this *Test) notifyUrl(ctx *gin.Context) {

	params := this.params(ctx)

	fmt.Println("==================== notifyUrl：", params)
}

// 测试网络请求
func (this *Test) request(ctx *gin.Context) {

	params := this.params(ctx)

	this.json(ctx, map[string]any{
		"method":  ctx.Request.Method,
		"params":  params,
		"headers": this.headers(ctx),
	}, facade.Lang(ctx, "数据请求成功！"), 200)
}
