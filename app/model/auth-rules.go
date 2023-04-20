package model

import (
	"gorm.io/plugin/soft_delete"
	"inis/app/facade"
)

type AuthRules struct {
	Id         int                   `gorm:"type:int(32); comment:主键;" json:"id"`
	Pid 	   int                   `gorm:"type:int(32); comment:父级ID; default:0;" json:"pid"`
	Name   	   string                `gorm:"comment:规则名称;" json:"name"`
	Method 	   string 				 `gorm:"comment:请求类型; default:'GET';" json:"method"`
	Route	   string				 `gorm:"comment:路由;" json:"route"`
	Common 	   int					 `gorm:"type:int(32); default:0; comment:是否公共规则;" json:"common"`
	Remark     string				 `gorm:"comment:备注; default:Null;" json:"remark"`
	CreateTime int64                 `gorm:"autoCreateTime; comment:创建时间; default:Null;" json:"create_time"`
	UpdateTime int64                 `gorm:"autoUpdateTime; comment:更新时间; default:Null;" json:"update_time"`
	DeleteTime soft_delete.DeletedAt `gorm:"comment:删除时间; default:0;" json:"delete_time"`
}

func init() {
	if migrate {
		err := facade.MySQL.Conn.AutoMigrate(&AuthRules{})
		if err != nil {
			facade.Log.Error("AuthRules表迁移失败", map[string]any{"error": err})
			return
		}
	}
}
