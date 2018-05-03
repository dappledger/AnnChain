package main

import (
	_ "github.com/dappledger/AnnChain/vanntool/models"
	_ "github.com/dappledger/AnnChain/vanntool/routers"

	"github.com/astaxie/beego"
)

func main() {
	beego.SetLogger("file", `{"filename":"logs/test.log"}`)
	beego.Run()
}
