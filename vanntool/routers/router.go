package routers

import (
	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/vanntool/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/cmdlist", &controllers.CmdListController{}, "get:Get;post:Post")
	beego.Router("/nodesop", &controllers.NodesOpController{}, "get:Get;post:Post")
	beego.Router("/nodeinfo", &controllers.ShowNodeInfo{}, "get:Get;post:Post")
	beego.Router("/util", &controllers.UtilController{}, "get:Get;post:Post")

	// for test
	beego.Router("/block", &controllers.GenBlockForTest{}, "get:Get")
	beego.Router("/last_height", &controllers.LastHeightForTest{}, "get:Get")
}
