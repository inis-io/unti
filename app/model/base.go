package model

import (
	"github.com/spf13/cast"
	"inis/app/facade"
)

var migrate bool

func init() {
	migrate = cast.ToBool(facade.DBToml.Get("mysql.auto_migrate"))
}