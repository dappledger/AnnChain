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
