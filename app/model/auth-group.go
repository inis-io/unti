package model

import (
	"github.com/spf13/cast"
	"gorm.io/plugin/soft_delete"
	"inis/app/facade"
)

type AuthGroup struct {
	Id         int                   `gorm:"type:int(32); comment:主键;" json:"id"`
	Name   	   string                `gorm:"comment:权限名称;" json:"name"`
	Rules 	   string 				 `gorm:"type:longtext; comment:权限规则;" json:"rules"`
	Uids	   string				 `gorm:"type:longtext; comment:用户ID;" json:"uids"`
	Remark     string				 `gorm:"comment:备注; default:Null;" json:"remark"`
	CreateTime int64                 `gorm:"autoCreateTime; comment:创建时间; default:Null;" json:"create_time"`
	UpdateTime int64                 `gorm:"autoUpdateTime; comment:更新时间; default:Null;" json:"update_time"`
	DeleteTime soft_delete.DeletedAt `gorm:"comment:删除时间; default:0;" json:"delete_time"`
}

func init() {
	if cast.ToBool(facade.NewToml("db").Get("mysql.auto_migrate")) {
		err := facade.NewDB("mysql").Drive().AutoMigrate(&AuthGroup{})
		if err != nil {
			facade.Log.Error(map[string]any{"error": err}, "AuthGroup表迁移失败")
			return
		}
	}
}
