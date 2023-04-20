package facade

import (
	"errors"
	AliYunClient "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	AliYunUtil "github.com/alibabacloud-go/openapi-util/service"
	AliYunUtilV2 "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	TencentCloud "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/unti-io/go-utils/utils"
	"gopkg.in/gomail.v2"
)

func init() {

	// 初始化配置文件
	initSMSToml()
	// 初始化缓存
	initSMS()

	// 监听配置文件变化
	SMSToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	SMSToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initSMS()
	})
}

// SMSToml - SMS配置文件
var SMSToml *utils.ViperResponse

// initSMSToml - 初始化SMS配置文件
func initSMSToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "sms",
		Content: `# ======== SMS 配置 ========

# 默认驱动
default    = "email"


# 邮件服务配置
[email]
# 邮件服务器地址
host     = "smtp.qq.com"
# 邮件服务端口
port     = 465
# 邮件账号
account  = "97783391@qq.com"
# 服务密码 - 不是邮箱密码
password = ""
# 邮件昵称
nickname = "兔子"
# 邮件签名
sign_name = "萌卜兔"


# 阿里云短信服务配置
[aliyun]
# 阿里云AccessKey ID
access_key_id 	  = ""
# 阿里云AccessKey Secret
access_key_secret = ""
# 阿里云短信服务endpoint
endpoint		  = "dysmsapi.aliyuncs.com"
# 短信签名
sign_name         = "萌卜兔"
# 验证码模板
verify_code       = ""


# 腾讯云短信服务配置
[tencent]
# 腾讯云COS SecretId
secret_id         = ""
# 腾讯云COS SecretKey
secret_key        = ""
# 腾讯云短信服务endpoint
endpoint          = "sms.tencentcloudapi.com"
# 腾讯云短信服务appid
sms_sdk_app_id	  = ""
# 短信签名
sign_name         = "萌卜兔"
# 验证码模板id
verify_code       = ""
# 区域
region            = "ap-guangzhou"
`,
	}).Read()

	if item.Error != nil {
		Log.Error("SMS配置初始化错误", map[string]any{
			"error": item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
		return
	}

	SMSToml = &item
}

// 初始化SMS
func initSMS() {

	port     := cast.ToInt(SMSToml.Get("email.port"))
	host     := cast.ToString(SMSToml.Get("email.host"))
	account  := cast.ToString(SMSToml.Get("email.account"))
	password := cast.ToString(SMSToml.Get("email.password"))
	GoMail   = &GoMailRequest{
		Client: gomail.NewDialer(host, port, account, password),
	}

	aliyunClient, err := AliYunClient.NewClient(&AliYunClient.Config{
		// 访问的域名
		Endpoint: tea.String(cast.ToString(SMSToml.Get("aliyun.endpoint", "dysmsapi.aliyuncs.com"))),
		// 必填，您的 AccessKey ID
		AccessKeyId: tea.String(cast.ToString(SMSToml.Get("aliyun.access_key_id"))),
		// 必填，您的 AccessKey Secret
		AccessKeySecret: tea.String(cast.ToString(SMSToml.Get("aliyun.access_key_secret"))),
	})

	if err != nil {
		Log.Error("阿里云短信服务初始化错误", map[string]any{
			"error": err,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
	}

	SMSAliYun = &AliYunSMS{
		Client: aliyunClient,
	}

	credential := common.NewCredential(
		cast.ToString(SMSToml.Get("tencent.secret_id")),
		cast.ToString(SMSToml.Get("tencent.secret_key")),
	)
	clientProfile := profile.NewClientProfile()
	clientProfile.HttpProfile.Endpoint = cast.ToString(SMSToml.Get("tencent.endpoint", "sms.tencentcloudapi.com"))
	tencentClient, err := TencentCloud.NewClient(
		credential,
		cast.ToString(SMSToml.Get("tencent.region", "ap-guangzhou")),
		clientProfile,
	)

	if err != nil {
		Log.Error("腾讯云短信服务初始化错误", map[string]any{
			"error": err,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
	}

	SMSTencent = &TencentSMS{
		Client: tencentClient,
	}

	switch cast.ToString(SMSToml.Get("default")) {
	case "email":
		SMS = GoMail
	case "aliyun":
		SMS = SMSAliYun
	case "tencent":
		SMS = SMSTencent
	default:
		SMS = GoMail
	}
}

// SMS - 短信
var SMS SMSInterface
var GoMail *GoMailRequest
// SMSAliYun - 阿里云短信
var SMSAliYun *AliYunSMS
// SMSTencent - 腾讯云短信
var SMSTencent *TencentSMS

// SMSInterface - 短信接口
type SMSInterface interface {
	// VerifyCode - 发送验证码：phone 手机号（必须），code 验证码（可选，不传则随机生成）
	VerifyCode(phone any, code ...any) (response *SMSResponse)
}

// SMSResponse - 短信响应
type SMSResponse struct {
	// 错误信息
	Error      error
	// 结果
	Result     any
	// 文本
	Text       string
	// 验证码
	VerifyCode string
}

// ================================== GoMail邮件服务 - 开始 ==================================

// GoMailRequest - GoMail邮件服务
type GoMailRequest struct {
	Client *gomail.Dialer
	// 邮件模板
	Template string
}

// VerifyCode - 发送验证码
func (this *GoMailRequest) VerifyCode(phone any, code ...any) (response *SMSResponse) {

	response = &SMSResponse{}

	if len(code) == 0 {
		code = append(code, utils.Rand.String(6, "0123456789"))
	}

	if utils.Is.Empty(this.Template) {
		this.Template = "您的验证码是：{code}，有效期5分钟。（打死也不要把验证码告诉别人）"
	}

	item     := gomail.NewMessage()
	nickname := cast.ToString(SMSToml.Get("email.nickname"))
	account  := cast.ToString(SMSToml.Get("email.account"))
	item.SetHeader("From", nickname + "<" + account + ">")
	// 发送给多个用户
	item.SetHeader("To", cast.ToString(phone))
	// 设置邮件主题
	item.SetHeader("Subject", cast.ToString(SMSToml.Get("email.sign_name")))
	// 替换验证码
	this.Template = utils.Replace(this.Template, map[string]any{
		"{code}": code[0],
	})
	// 设置邮件正文
	item.SetBody("text/html", this.Template)

	// 发送邮件
	err := this.Client.DialAndSend(item)

	if err != nil {
		response.Error = err
		return response
	}

	response.VerifyCode = cast.ToString(code[0])

	return response
}

// ================================== 阿里云短信 - 开始 ==================================

// AliYunSMS - 阿里云短信
type AliYunSMS struct {
	Client *AliYunClient.Client
}

// VerifyCode - 发送验证码
func (this *AliYunSMS) VerifyCode(phone any, code ...any) (response *SMSResponse) {

	response = &SMSResponse{}

	params := map[string]any{
		// 必填，接收短信的手机号码
		"PhoneNumbers":  tea.String(cast.ToString(phone)),
		// 必填，短信签名名称
		"SignName":      tea.String(cast.ToString(SMSToml.Get("aliyun.sign_name"))),
		// 必填，短信模板ID
		"TemplateCode":  tea.String(cast.ToString(SMSToml.Get("aliyun.verify_code"))),
	}

	if len(code) == 0 {
		code = append(code, utils.Rand.String(6, "0123456789"))
	}

	params["TemplateParam"] = tea.String(utils.Json.Encode(map[string]any{
		"code": code[0],
	}))

	runtime := &AliYunUtilV2.RuntimeOptions{}
	request := &AliYunClient.OpenApiRequest{
		Query: AliYunUtil.Query(params),
	}

	// 返回值为 Map 类型，可从 Map 中获得三类数据：响应体 body、响应头 headers、HTTP 返回的状态码 statusCode
	result, err := this.Client.CallApi(this.ApiInfo(), request, runtime)
	if err != nil {
		response.Error = err
		return response
	}

	body := cast.ToStringMap(result["body"])
	if body["Code"] != "OK" {
		response.Error = errors.New(cast.ToString(body["Message"]))
		return response
	}

	response.Result 	= result
	response.Text 		= utils.Json.Encode(result)
	response.VerifyCode = cast.ToString(code[0])

	return response
}

// ApiInfo - 接口信息
func (this *AliYunSMS) ApiInfo() (result *AliYunClient.Params) {
	return &AliYunClient.Params{
		// 接口名称
		Action  : tea.String("SendSms"),
		// 接口版本
		Version : tea.String("2017-05-25"),
		// 接口协议
		Protocol: tea.String("HTTPS"),
		// 接口 HTTP 方法
		Method  : tea.String("POST"),
		AuthType: tea.String("AK"),
		Style   : tea.String("RPC"),
		// 接口 PATH
		Pathname: tea.String("/"),
		// 接口请求体内容格式
		ReqBodyType: tea.String("json"),
		// 接口响应体内容格式
		BodyType   : tea.String("json"),
	}
}

// ================================== 腾讯云短信 - 开始 ==================================

// TencentSMS - 腾讯云短信
type TencentSMS struct {
	Client *TencentCloud.Client
}

// VerifyCode - 发送验证码
func (this *TencentSMS) VerifyCode(phone any, code ...any) (response *SMSResponse) {

	response = &SMSResponse{}

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := TencentCloud.NewSendSmsRequest()

	if len(code) == 0 {
		code = append(code, utils.Rand.String(6, "0123456789"))
	}

	request.PhoneNumberSet   = common.StringPtrs([]string{cast.ToString(phone)})
	request.SmsSdkAppId      = common.StringPtr(cast.ToString(SMSToml.Get("tencent.sms_sdk_app_id")))
	request.SignName         = common.StringPtr(cast.ToString(SMSToml.Get("tencent.sign_name")))
	request.TemplateId       = common.StringPtr(cast.ToString(SMSToml.Get("tencent.verify_code")))
	request.TemplateParamSet = common.StringPtrs([]string{cast.ToString(code[0])})

	item, err := this.Client.SendSms(request)

	if err != nil {
		response.Error = err
		return response
	}

	if item.Response == nil {
		response.Error = errors.New("response is nil")
		return response
	}

	if len(item.Response.SendStatusSet) == 0 {
		response.Error = errors.New("response send status set is nil")
		return response
	}

	if *item.Response.SendStatusSet[0].Code != "Ok" {
		response.Error = errors.New(cast.ToString(item.Response.SendStatusSet[0].Message))
		return response
	}

	response.VerifyCode = cast.ToString(code[0])
	response.Text       = item.ToJsonString()
	response.Result     = utils.Json.Decode(item.ToJsonString())

	return response
}