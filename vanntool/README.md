
go get github.com/astaxie/beego

cd $GOPATH/src/github.com/dappledger/AnnChain/

git clone git@gitlab.zhonganonline.com:ann/vanntool.git

cd vanntool

## 配置anntool包
方案1: cp $GOPATH/src/github.com/dappledger/AnnChain//build/anntool ./bin/anntool

方案2: conf/config.json，添加
```json
{
	"AnntoolPath":"/home/root/workspace/go/src/github.com/dappledger/AnnChain//build/anntool"
}
```

bee run # or `go build main.go`

## pack
bee pack -be -v -exp=vendor GOOS=linux
