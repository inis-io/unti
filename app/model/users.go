package model

import (
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"inis/app/facade"
	regexp2 "regexp"
)

type Users struct {
	Id          int    `gorm:"type:int(32); comment:主键;" json:"id"`
	Account     string `gorm:"size:32; comment:帐号; default:Null;" json:"account"`
	Password    string `gorm:"comment:密码;" json:"password"`
	Nickname    string `gorm:"size:32; comment:昵称;" json:"nickname"`
	Email       string `gorm:"size:128; comment:邮箱;" json:"email"`
	Phone       string `gorm:"size:32; comment:手机号;" json:"phone"`
	Avatar      string `gorm:"comment:头像; default:Null;" json:"avatar"`
	Description string `gorm:"comment:描述; default:Null;" json:"description"`
	Title       string `gorm:"comment:头衔; default:Null;" json:"title"`
	Exp  		int    `gorm:"type:int(32); comment:经验值; default:0;" json:"exp"`
	Pages       string `gorm:"comment:页面权限; default:Null;" json:"pages"`
	Source      string `gorm:"size:32; default:'default'; comment:注册来源;" json:"source"`
	Remark      string `gorm:"comment:备注; default:Null;" json:"remark"`
	// 以下为公共字段
	Json       any                   `gorm:"type:longtext; comment:用于存储JSON数据;" json:"json"`
	Text       any                   `gorm:"type:longtext; comment:用于存储文本数据;" json:"text"`
	Result     any                   `gorm:"type:varchar(256); comment:不存储数据，用于封装返回结果;" json:"result"`
	LoginTime  int64                 `gorm:"size:32; comment:登录时间; default:Null;" json:"login_time"`
	CreateTime int64                 `gorm:"autoCreateTime; comment:创建时间;" json:"create_time"`
	UpdateTime int64                 `gorm:"autoUpdateTime; comment:更新时间;" json:"update_time"`
	DeleteTime soft_delete.DeletedAt `gorm:"comment:删除时间; default:0;" json:"delete_time"`
}

// InitUsers - 初始化Users表
func InitUsers() {
	// 迁移表
	err := facade.DB.Drive().AutoMigrate(&Users{})
	if err != nil {
		facade.Log.Error(map[string]any{"error": err}, "Users表迁移失败")
		return
	}
}

// AfterFind - 查询后的钩子
func (this *Users) AfterFind(tx *gorm.DB) (err error) {

	if utils.Is.Empty(this.Avatar) {

		// 正则匹配邮箱 [1-9]\d+@qq.com 是否匹配
		reg := regexp2.MustCompile(`[1-9]\d+@qq.com`).MatchString(this.Email)
		if reg {

			// 获取QQ号
			qq := regexp2.MustCompile(`[1-9]\d+`).FindString(this.Email)
			this.Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + qq + "&s=100"

		} else {
			avatars := utils.File(utils.FileRequest{
				Ext:    ".png, .jpg, .jpeg, .gif",
				Dir:    "public/assets/images/avatar/",
				Domain: fmt.Sprintf("%v/", facade.Cache.Get("domain")),
				Prefix: "public/",
			}).List()

			// 随机获取头像
			if len(avatars.Slice) > 0 {
				this.Avatar = cast.ToString(avatars.Slice[utils.Rand.Int(0, len(avatars.Slice)-1)])
			}
		}
	}

	// 替换 url 中的域名
	this.Avatar = utils.Replace(this.Avatar, DomainTemp1())

	this.Result = map[string]any{

	}
	this.Text = cast.ToString(this.Text)
	this.Json = utils.Json.Decode(this.Json)

	return
}

// AfterSave - 保存后的Hook（包括 create update）
func (this *Users) AfterSave(tx *gorm.DB) (err error) {

	go func() {
		this.Avatar = utils.Replace(this.Avatar, DomainTemp2())
		tx.Model(this).UpdateColumn("avatar", this.Avatar)
	}()

	// 账号 唯一处理
	if !utils.Is.Empty(this.Account) {
		exist := facade.DB.Model(&Users{}).Where("id", "!=", this.Id).Where("account", this.Account).Exist()
		if exist {
			return errors.New("账号已存在！")
		}
	}

	// 邮箱 唯一处理
	if !utils.Is.Empty(this.Email) {
		exist := facade.DB.Model(&Users{}).Where("id", "!=", this.Id).Where("email", this.Email).Exist()
		if exist {
			return errors.New("邮箱已存在！")
		}
	}

	// 手机号 唯一处理
	if !utils.Is.Empty(this.Phone) {
		exist := facade.DB.Model(&Users{}).Where("id", "!=", this.Id).Where("phone", this.Phone).Exist()
		if exist {
			return errors.New("手机号已存在！")
		}
	}

	return
}