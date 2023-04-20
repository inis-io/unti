package model

import (
	"fmt"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"inis/app/facade"
	regexp2 "regexp"
)

type Users struct {
	Id          int                   `gorm:"type:int(32); comment:主键;" json:"id"`
	Account     string                `gorm:"size:32; comment:帐号; default:Null;" json:"account"`
	Password    string                `gorm:"comment:密码;" json:"password"`
	Nickname    string                `gorm:"size:32; comment:昵称;" json:"nickname"`
	Email       string                `gorm:"size:128; comment:邮箱;" json:"email"`
	Avatar      string                `gorm:"comment:头像; default:Null;" json:"avatar"`
	Description string                `gorm:"comment:描述; default:Null;" json:"description"`
	Level       string                `gorm:"size:32; default:'user'; comment:权限;" json:"level"`
	Source      string                `gorm:"size:32; default:'default'; comment:注册来源;" json:"source"`
	Json 	  	any                	  `gorm:"type:longtext; comment:用于存储JSON数据;" json:"json"`
	Result 		any				  	  `gorm:"type:varchar(256); comment:不存储数据，用于封装返回结果;" json:"result"`
	LoginTime   int64                 `gorm:"size:32; comment:登录时间; default:Null;" json:"login_time"`
	CreateTime  int64                 `gorm:"autoCreateTime; comment:创建时间;" json:"create_time"`
	UpdateTime  int64                 `gorm:"autoUpdateTime; comment:更新时间;" json:"update_time"`
	DeleteTime  soft_delete.DeletedAt `gorm:"comment:删除时间; default:0;" json:"delete_time"`
}

func init() {
	if migrate {
		err := facade.MySQL.Conn.AutoMigrate(&Users{})
		if err != nil {
			facade.Log.Error("Users表迁移失败", map[string]any{"error": err})
			return
		}
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
				Ext: ".png, .jpg, .jpeg, .gif",
				Dir: "public/assets/images/avatar/",
				Domain: fmt.Sprintf("%v/", "https://inis.unti.io"),
				Prefix: "public/",
			}).List()

			// 随机获取头像
			if len(avatars.Slice) > 0 {
				this.Avatar = cast.ToString(avatars.Slice[utils.Rand.Int(0, len(avatars.Slice)-1)])
			}
		}
	} else {
		// 判断 this.Avatar 中是否包含 %s 字符
		reg := regexp2.MustCompile(`%s`).MatchString(this.Avatar)
		this.Avatar = utils.Ternary(reg, fmt.Sprintf(this.Avatar, "https://inis.unti.io"), this.Avatar)
	}

	this.Result = utils.JsonDecode(utils.JsonEncode(map[string]any{
		"id": this.Id,
		"array": []any{1, 2, 3, 4, 5},
		"map": map[string]any{
			"key": "value",
			"key2": "value2",
		},
		"string": "这是封装的拓展字段",
	}))

	return
}
