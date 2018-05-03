package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

func init() {
	config.init()
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", config.GetDataPath())
	orm.RegisterModel(new(NodeData))
	err := orm.RunSyncdb("default", false, true)
	if err != nil {
		beego.Error("[init_config],err:", err)
	}

	dm.init()
}
