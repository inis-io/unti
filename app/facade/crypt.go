package facade

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	JWT "github.com/dgrijalva/jwt-go"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"time"
)

// AppToml - APP配置文件
var AppToml *utils.ViperResponse

// initAppToml - 初始化APP配置文件
func initAppToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "app",
		Content: `# ======== 基础服务配置 - 修改此文件建议重启服务 ========

# 应用配置
[app]
# 项目运行端口
port        = 1000
# 调试模式
debug       = false
# 版本号
version     = "2.0.0"
# 登录token名称（别乱改，别作死）
token_name  = "unti_login_token"

# JWT加密配置
[jwt]
# jwt密钥
secret   = "unti_api_key"
# 过期时间(秒)
expire   = 604800

# API限流器配置
[qps]
# 单个接口每秒最大请求数
point    = 10
# 全局接口每秒最大请求数
global   = 50
`,
	}).Read()

	if item.Error != nil {
		Log.Error(map[string]any{
			"error": item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "APP配置初始化错误")
		return
	}

	AppToml = &item
}

// 初始化缓存
func initApp() {

}

func init() {
	// 初始化配置文件
	initAppToml()
	// 初始化缓存
	initApp()

	// 监听配置文件变化
	AppToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	AppToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initApp()
	})

	Jwt.Create = JwtCreate
	Jwt.Parse  = JwtParse
}

var Jwt struct {
	Create func(data map[string]any) (result string, error error)
	Parse  func(token string) (result JwtStruct)
}

type configJwt struct {
	Data map[string]any `json:"data"`
	JWT.StandardClaims
}

type JwtStruct struct {
	Data  map[string]any `json:"data"`
	Error error `json:"error"`
	Valid int64 `json:"valid"`
}

// JwtCreate 创建token
func JwtCreate(data map[string]any) (result string, err error) {
	return JWT.NewWithClaims(JWT.SigningMethodHS256, configJwt{
		Data: data,
		StandardClaims: JWT.StandardClaims{
			ExpiresAt:  time.Now().Unix() + cast.ToInt64(AppToml.Get("jwt.expire", "7200")), // 过期时间戳
			IssuedAt:   time.Now().Unix(),                                                   		   // 当前时间戳
			Issuer:     cast.ToString(AppToml.Get("jwt.issuer", "unti")),                    // 颁发者签名
			Subject:    cast.ToString(AppToml.Get("jwt.subject", "unti-io")),                // 签名主题
		},
	}).SignedString([]byte(cast.ToString(AppToml.Get("jwt.secret", "inis"))))
}

// JwtParse 解析token
func JwtParse(token string) (result JwtStruct) {

	item, err := JWT.ParseWithClaims(token, &configJwt{}, func(token *JWT.Token) (any, error) {
		return []byte(cast.ToString(AppToml.Get("jwt.secret", "inis"))), nil
	})

	if err != nil {
		Log.Error(map[string]any{
			"error": err,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "JWT解析错误")
		result.Error = err
		return
	}

	if key, _ := item.Claims.(*configJwt); item.Valid {
		result.Data  = key.Data
		result.Valid = key.StandardClaims.ExpiresAt - time.Now().Unix()
	}

	return
}

// CipherRequest - 请求输入
type CipherRequest struct {
	// 16位密钥
	Key string
	// 16位向量
	Iv  string
}

// CipherResponse - 响应输出
type CipherResponse struct {
	// 加密后的字节
	Byte []byte
	// 加密后的字符串
	Text string
	// 错误信息
	Error error
}

// Cipher - 对称加密
func Cipher(key, iv any) *CipherRequest {
	return &CipherRequest{
		Key: cast.ToString(key),
		Iv:  cast.ToString(iv),
	}
}

// Encrypt 加密
func (this *CipherRequest) Encrypt(text any) (result *CipherResponse) {

	result = &CipherResponse{}

	// 拦截异常
    defer func() {
        if r := recover(); r != nil {
            result.Error = fmt.Errorf("%v", r)
        }
    }()

	block, err := aes.NewCipher([]byte(this.Key))
	if err != nil {
		result.Error = err
	}

	// 每个块的大小
	blockSize := block.BlockSize()
	// 计算需要填充的长度
	padding   := blockSize - len([]byte(cast.ToString(text))) % blockSize

	// 填充
	fill   := append([]byte(cast.ToString(text)), bytes.Repeat([]byte{byte(padding)}, padding)...)
	encode := make([]byte, len(fill))

	item := cipher.NewCBCEncrypter(block, []byte(this.Iv))
	item.CryptBlocks(encode, fill)

	result.Byte = encode
	result.Text = base64.StdEncoding.EncodeToString(encode)

	return
}

// Decrypt 解密
func (this *CipherRequest) Decrypt(text any) (result *CipherResponse) {

    result = &CipherResponse{}

	// 拦截异常
    defer func() {
        if r := recover(); r != nil {
            result.Error = fmt.Errorf("%v", r)
        }
    }()

    newText, err := base64.StdEncoding.DecodeString(cast.ToString(text))
    if err != nil {
        result.Error = err
        return
    }

    block, err := aes.NewCipher([]byte(this.Key))
    if err != nil {
        result.Error = err
        return
    }

    // 确保 newText 是 blockSize 的整数倍
    blockSize := block.BlockSize()
    if len(newText) % blockSize != 0 {
        result.Error = errors.New("invalid ciphertext")
        return
    }

    decode := make([]byte, len(newText))
    item := cipher.NewCBCDecrypter(block, []byte(this.Iv))
    item.CryptBlocks(decode, newText)

    // 去除填充
    padding := decode[len(decode)-1]
    result.Byte = decode[:len(decode)-int(padding)]
    result.Text = string(result.Byte)

    return
}