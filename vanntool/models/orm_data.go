package models

type NodeData struct {
	Name    string `orm:"unique;pk"`
	RPCAddr string `orm:"column(rpc_addr);unique"`
	IP      string `orm:"column(ip)"`
	Privkey string
}

type NodeDataShow struct {
	Name    string
	RPCAddr string
}

func NodeSlcToShowSlc(allData []*NodeData) []NodeDataShow {
	ret := make([]NodeDataShow, len(allData))
	for i := range allData {
		ret[i].Name = allData[i].Name
		ret[i].RPCAddr = allData[i].RPCAddr
	}
	return ret
}
