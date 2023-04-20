package facade

import (
	"github.com/go-pay/gopay/alipay"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"os"
)

func init() {
	// 初始化配置文件
	initPayToml()
}

// PayToml - 支付配置文件
var PayToml *utils.ViperResponse

// initCacheToml - 初始化缓存配置文件
func initPayToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "pay",
		Content: `# ======== 支付配置 ========

# 支付宝支付
[alipay]
# 支付宝支付的商户ID
app_id                 = 20210***28
# 证书根目录
root_cert_path         = "/config/pay/ali/"
# 应用私钥
app_private_key_path   = "appPrivateKey.pem"
# 支付宝公钥
alipay_public_key_path = "alipayPublicKey.pem"
# 异步通知地址
notify_url = "https://api.inis.cn/api/test/notify"
# 同步通知地址
return_url = "https://api.inis.cn/api/test/return"
# 时区
time_zone = "Asia/Shanghai"
`,
	}).Read()

	if item.Error != nil {
		Log.Error("支付配置初始化错误", map[string]any{
			"error": item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
		return
	}

	PayToml = &item
}

var pay struct {
	Alipay func() *alipay.Client
}

// Alipay - 支付宝支付
func Alipay() *alipay.Client {

	// 当前目录
	path, _ := os.Getwd()
	// 证书路径
	path = path + cast.ToString(PayToml.Get("alipay.root_cert_path", "/config/pay/ali/"))
	// 应用私钥
	privateKey := utils.File().Byte(path + cast.ToString(PayToml.Get("alipay.app_private_key_path", "appPrivateKey.pem")))

	// 初始化支付宝客户端
	//    appid：应用ID
	//    privateKey：应用私钥，支持PKCS1和PKCS8
	//    isProd：是否是正式环境
	client, err := alipay.NewClient(cast.ToString(PayToml.Get("alipay.app_id")), privateKey.Text, true)
	if err != nil {
		Log.Error("支付宝支付初始化失败", map[string]any{
			"err": err,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
		return nil
	}

	// 设置支付宝请求 公共参数
	//    注意：具体设置哪些参数，根据不同的方法而不同，此处列举出所有设置参数
	// 设置时区，不设置或出错均为默认服务器时间
	client.SetLocation(cast.ToString(PayToml.Get("alipay.time_zone", alipay.LocationShanghai)))
	// 设置返回URL
	client.SetReturnUrl(cast.ToString(PayToml.Get("alipay.return_url")))
	// 设置异步通知URL
	client.SetNotifyUrl(cast.ToString(PayToml.Get("alipay.notify_url")))

	return client
}