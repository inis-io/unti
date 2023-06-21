package facade

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	JWT "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"hash/fnv"
	"time"
)

// CryptToml - Crypt配置文件
var CryptToml *utils.ViperResponse

// initCryptToml - 初始化Crypt配置文件
func initCryptToml() {

	key := fmt.Sprintf("%v-%v", uuid.New().String(), time.Now().Unix())
	secret := fmt.Sprintf("Unti-%x", md5.Sum([]byte(key)))

	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "crypt",
		Content: utils.Replace(TempCrypt, map[string]any{
			"${jwt.key}": 		secret,
			"${jwt.expire}":    "7 * 24 * 60 * 60",
			"${jwt.issuer}" :   "unti.io",
			"${jwt.subject}":   "Unti",
		}),
	}).Read()

	if item.Error != nil {
		Log.Error(map[string]any{
			"error":     item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "Crypt配置初始化错误")
		return
	}

	CryptToml = &item
}

// 初始化加密配置
func initCrypt() {

}

func init() {
	// 初始化配置文件
	initCryptToml()
	// 初始化缓存
	initCrypt()

	// 监听配置文件变化
	CryptToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	CryptToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initCrypt()
	})
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
		request[0].Expire = cast.ToInt64(utils.Calc(CryptToml.Get("jwt.expire", "7200")))
	}

	// 颁发者签名
	if utils.Is.Empty(request[0].Issuer) {
		request[0].Issuer = cast.ToString(CryptToml.Get("jwt.issuer", "unti.io"))
	}

	// 主题
	if utils.Is.Empty(request[0].Subject) {
		request[0].Subject = cast.ToString(CryptToml.Get("jwt.subject", "Unti"))
	}

	// 密钥
	if utils.Is.Empty(request[0].Key) {
		request[0].Key = cast.ToString(CryptToml.Get("jwt.key", "Unti"))
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
		Log.Error(map[string]any{
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

// CipherRequest - 请求输入
type CipherRequest struct {
	// 16位密钥
	Key string
	// 16位向量
	Iv string
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
	padding := blockSize - len([]byte(cast.ToString(text)))%blockSize

	// 填充
	fill := append([]byte(cast.ToString(text)), bytes.Repeat([]byte{byte(padding)}, padding)...)
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
	if len(newText)%blockSize != 0 {
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

type HashStruct struct {}

// Hash - 哈希加密
var Hash = &HashStruct{}

// Sum32 - 哈希加密
func (this *HashStruct) Sum32(text any) (result string) {
	item := fnv.New32()
	_, err := item.Write([]byte(cast.ToString(text)))
	return cast.ToString(utils.Ternary[any](err != nil, nil, item.Sum32()))
}