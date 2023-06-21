package facade

// TempApp - APP配置模板
const TempApp = `# ======== 基础服务配置 - 修改此文件建议重启服务 ========

# 应用配置
[app]
# 项目运行端口
port        = 8642
# 调试模式
debug       = false
# 登录token名称（别乱改，别作死）
token_name  = "UNTI_LOGIN_TOKEN"
`

// TempDatabase - 数据库配置模板
const TempDatabase = `# ======== 数据库配置 ========

# 默认数据库配置
default    = "mysql"

# mysql 数据库配置
[mysql]
# 数据库类型
type         = "mysql"
# 数据库地址
hostname     = "${mysql.hostname}"
# 数据库端口
hostport     = ${mysql.hostport}
# 数据库用户
username     = "${mysql.username}"
# 数据库名称
database     = "${mysql.database}"
# 数据库密码
password     = "${mysql.password}"
# 数据库编码
charset      = "${mysql.charset}"
# 表前缀
prefix       = "unti_"
# 自动迁移模式
migrate 	 = ${mysql.migrate}
`

// TempCache - 缓存配置模板
const TempCache = `# ======== 缓存配置 ========

# 是否开启API缓存
open	   = ${open}
# 默认缓存驱动
default    = "${default}"

# redis配置
[redis]
# redis地址
host       = "${redis.host}"
# redis端口
port       = ${redis.port}
# redis密码
password   = "${redis.password}"
# 过期时间(秒) - 0为永不过期
expire     = "${redis.expire}"
# redis前缀
prefix     = "${redis.prefix}"
# redis数据库
database   = ${redis.database}

# 文件缓存配置
[file]
# 缓存过期时间(秒) - 0为永不过期
expire     = "${file.expire}"
# 缓存目录
path       = "${file.path}"
# 缓存前缀
prefix     = "${file.prefix}"

# 内存缓存配置
[ram]
# 缓存过期时间(秒) - 0为永不过期
expire     = "${ram.expire}"
`

// TempLog - 日志配置模板
const TempLog = `# ======== 日志配置 ========

# 是否启用日志
on		   = ${on}
# 单个日志文件大小（MB）
size	   = ${size}
# 日志文件保存天数
age		   = ${age}
# 日志文件最大保存数量
backups	   = ${backups}
`

// TempPay - 支付配置模板
const TempPay = `# ======== 支付配置 ========

# 支付宝支付
[alipay]
# 支付宝支付的商户ID
app_id                 = ${alipay.app_id}
# 证书根目录
root_cert_path         = "${alipay.root_cert_path}"
# 应用私钥
app_private_key_path   = "${alipay.app_private_key_path}"
# 支付宝公钥
alipay_public_key_path = "${alipay.alipay_public_key_path}"
# 异步通知地址
notify_url = "${alipay.notify_url}"
# 同步通知地址
return_url = "${alipay.return_url}"
# 时区
time_zone  = "${alipay.time_zone}"
`

// TempSMS - 短信配置模板
const TempSMS = `# ======== SMS 配置 ========

# 驱动
[drive]
# 邮件
email     = "${drive.email}"
# 短信
sms	      = "${drive.sms}"
# 默认
default   = "${drive.default}"


# 邮件服务配置
[email]
# 邮件服务器地址
host      = "${email.host}"
# 邮件服务端口
port      = ${email.port}
# 邮件账号
account   = "${email.account}"
# 服务密码 - 不是邮箱密码
password  = "${email.password}"
# 邮件昵称
nickname  = "${email.nickname}"
# 邮件签名
sign_name = "${email.sign_name}"


# 阿里云短信服务配置
[aliyun]
# 阿里云AccessKey ID
access_key_id 	  = "${aliyun.access_key_id}"
# 阿里云AccessKey Secret
access_key_secret = "${aliyun.access_key_secret}"
# 阿里云短信服务endpoint
endpoint		  = "${aliyun.endpoint}"
# 短信签名
sign_name         = "${aliyun.sign_name}"
# 验证码模板
verify_code       = "${aliyun.verify_code}"


# 腾讯云短信服务配置
[tencent]
# 腾讯云SecretId
secret_id         = "${tencent.secret_id}"
# 腾讯云SecretKey
secret_key        = "${tencent.secret_key}"
# 腾讯云短信服务endpoint
endpoint          = "${tencent.endpoint}"
# 腾讯云短信服务appid
sms_sdk_app_id	  = "${tencent.sms_sdk_app_id}"
# 短信签名
sign_name         = "${tencent.sign_name}"
# 验证码模板id
verify_code       = "${tencent.verify_code}"
# 区域
region            = "${tencent.region}"
`

// TempStorage - 存储配置模板
const TempStorage = `# ======== 存储配置 ========

# 默认存储驱动
default    = "${default}"


# 本地存储配置
[local]
# 本地存储域名
domain     = "${local.domain}"


# 阿里OSS配置
[oss]
# 阿里云AccessKey ID
access_key_id 	  = "${oss.access_key_id}"
# 阿里云AccessKey Secret
access_key_secret = "${oss.access_key_secret}"
# OSS 外网 Endpoint
endpoint		  = "${oss.endpoint}"
# OSS Bucket - 存储桶名称
bucket			  = "${oss.bucket}"
# OSS 外网域名 - 用于访问 - 不填写则使用默认域名
domain			  = "${oss.domain}"


# 腾讯云COS配置
[cos]
# 腾讯云COS AppId
app_id            = "${cos.app_id}"
# 腾讯云COS SecretId
secret_id         = "${cos.secret_id}"
# 腾讯云COS SecretKey
secret_key        = "${cos.secret_key}"
# COS Bucket - 存储桶名称
bucket            = "${cos.bucket}"
# COS 所在地区，如这里的 ap-guangzhou（广州）
region            = "${cos.region}"
# COS 外网域名 - 用于访问 - 不填写则使用默认域名
domain            = "${cos.domain}"


# 七牛云KODO配置
[kodo]
# 七牛云AccessKey
access_key        = "${kodo.access_key}"
# 七牛云SecretKey
secret_key        = "${kodo.secret_key}"
# KODO Bucket - 存储桶名称
bucket            = "${kodo.bucket}"
# KODO 所在地区，如这里的华南（广东） z0=华东 z1=华北河北 z2=华南广东 cn-east-2=华东浙江 na0=北美 as0=新加坡 ap-northeast-1=亚太-首尔机房
region            = "${kodo.region}"
# KODO 外网域名 - 用于访问 - 这里必须填写
domain            = "${kodo.domain}"
`

const TempCrypt   = `# ======== 加密配置 ========

# JWT配置
[jwt]
# 密钥
key      = "${jwt.key}"
# 过期时间(秒)
expire   = "${jwt.expire}"
# 签发者
issuer   = "${jwt.issuer}"
# 主题
subject  = "${jwt.subject}"
`