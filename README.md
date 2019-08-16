## How to build?

``` shell
# Step0. 设置GOPATH。
export GOPATH=$HOME/.gopkgs

# Step1. Clone项目到本地, 注意不要放到GOPATH中。另外，这里使用的是http方式，如果clone不了，请参考后面的解决方案。
git clone https://github.com/dappledger/AnnChain.git

# Step2. 进入项目，下载依赖。
cd AnnChain
./get_pkgs.sh

# Step3. 构建。
make
```

## How to run?

### use docker compose to run in one host.


```
## start cluster
➜  docker-compose up

## remove cluster
➜  docker-compose down
```

##设置validator运行的节点
并不是所有的k8s 节点都会运行genesis。 通过gtool生成的集群配置中，会自动给每个genesis分配nodeSelector: `validatorNode<i>=enabled`,所以，我们需要给k8s node添加标签 `validator<i>=enabled`
具体操作示例如下：
```bash
kubectl label node <node_name> validatorNode7=enabled
```
