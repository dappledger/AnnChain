package routers

import (
	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/src/vision/controllers"
)

func InitNode(runtime string) {
	beego.Router("/", &controllers.InitNode{
		Runtime: runtime,
	}, "get:Get;post:Post")
}
