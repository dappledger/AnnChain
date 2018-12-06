#!/bin/bash

##########################################################################
# This is the annChain.Genesis automated install script for Linux and Mac OS.
# This file was downloaded from https://github.com/dappledger/AnnChain
#
# Copyright (c) 2018, Respective Authors all rights reserved.
#
# After December 1, 2018 this software is available under the following terms:
# 
# The MIT License
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#
# https://github.com/dappledger/AnnChain/blob/master/README.md
##########################################################################

VERSION=0.1
TIME_BEGIN=$( date -u +%s )

#install golang
function InstallGo()
{
    source /etc/os-release
    case $ID in
    debian|ubuntu|devuan)
        sudo apt-get install curl
        ;;
    centos|fedora|rhel)
        yumdnf="yum"
        if test "$(echo "$VERSION_ID >= 22" | bc)" -ne 0;then
            yumdnf="dnf"
        fi
        $yumdnf install -y curl
        ;;
    *)
        exit 1
        ;;
    esac

    echo -e "Please enter the genesis requirements above go 1.9:\c"
    read -e go_version

    curl -SL -o go.tar.gz https://www.golangtc.com/static/go/${go_version}/go${go_version}.linux-amd64.tar.gz && \
    tar -C  /usr/local/ -zxf go.tar.gz && \
    rm -rf go.tar.gz
    
    #GOPATH
    echo -e "Please enter the GOPATH dir:\c"
    read -e GOPATHDIR
    mkdir -p ${GOPATHDIR}
#config env
cat <<EOF >> /etc/profile
export GOROOT=/usr/local/go
export GOPATH=${GOPATHDIR}
export PATH=\$GOROOT/bin:\$PATH
EOF
    source /etc/profile
    go env
    go version
}

#install git 
function InstallGit()
{
    source /etc/os-release
    case $ID in
    debian|ubuntu|devuan)
        sudo apt-get install git
        ;;
    centos|fedora|rhel)
        yumdnf="yum"
        if test "$(echo "$VERSION_ID >= 22" | bc)" -ne 0;then
            yumdnf="dnf"
        fi
        $yumdnf install -y git
        ;;
    *)
        exit 1
        ;;
    esac 
}

#Determine if the command was executed successfully
function ExIsOk()
{
    if [ $? -ne 0 ];then
        echo "Command execution failed"
        exit 1
    fi
}

#build Genesis
function BuildGenesis()
{
    git clone http://github.com/dappledger/AnnChain/ann-module.git ${GOPATH}/src/github.com/dappledger/AnnChain/ann-module
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/ann-module
    git checkout -b genesis origin/genesis >/dev/null 2>&1
    git clone http://github.com/dappledger/AnnChain/angine.git ${GOPATH}/src/github.com/dappledger/AnnChain/angine
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/angine
    git checkout -b genesis origin/genesis >/dev/null 2>&1
    git clone http://github.com/dappledger/AnnChain/go-sdk.git ${GOPATH}/src/github.com/dappledger/AnnChain/go-sdk
    git clone http://github.com/dappledger/AnnChain/genesis.git ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
    git checkout -b genesis origin/genesis >/dev/null 2>&1
    printf "\\n\\tBeginning build version: %s\\n" "${VERSION}"
    make
}

#init Genesis
function InitGenesis()
{
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis/build
    ./genesis init >/dev/null 2>&1 &
}

#Configuration config.toml
function ConfigToml()
{
    GENESIS_PATH=$HOME/.ann_runtime
    echo -e "Please enter the Chain-ID:\c"
    read -e CHAINID
    sed -i 's/"annchain.*"/"'${CHAINID}'"/g' ${GENESIS_PATH}/genesis.json

    echo -e "Please enter the P2P Listen on port:\c"
    read -e NODE_LADDR
    sed -i "s/46656/${NODE_LADDR}/g" ${GENESIS_PATH}/config.toml

    echo -e "Please enter the RPC Listen on port:\c"
    read -e RPC_LADDR
    sed -i "s/46657/${RPC_LADDR}/g" ${GENESIS_PATH}/config.toml

    echo -e "Please enter the SEEDS P2P NODE format(IP:PORT):\c"
    read -e SEEDS
    sed -i 's/seeds = ""/seeds = "'${SEEDS}'"/g' ${GENESIS_PATH}/config.toml
    
}

#Run Genesis
function RunGenesis()
{
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis/build
    nohup ./genesis node >/dev/null 2>&1  &
}

#Check that the Genesis process is successful
function CheckGenesisIsOK()
{
    txtbld=$(tput bold)
    bldred=${txtbld}$(tput setaf 2)
    txtrst=$(tput sgr0)
    TIME_END=$(( $(date -u +%s) - ${TIME_BEGIN} ))
    SERVICE_STATUS=`ps -ef | grep genesis | grep -v grep|wc -l`
    if [ $SERVICE_STATUS -gt 0 ];then
	#printf "\t#####################################################\v"
	printf "${bldred}\n"
        printf "\t     _                  ____            _     _\n"     
	printf "\t    / \    |\    |\    |    \|    |    / \    |\    |\n"
	printf "\t   /   \   | \   | \   |     |    |   /   \   | \   |\n"
	printf "\t  / ___ \  |  \  |  \  |     |____|  /_____\  |  \  |\n"
	printf "\t /       \ |   \ |   \ |     |    | /       \ |   \ |\n"
	printf "\t/         \|    \|    \|____/|    |/         \|    \|\n${txtrst}\n"
	#printf "\t#####################################################\v"
	printf "\\n\\tGenesis has been successfully built. %02d:%02d:%02d\\n\\n" $(($TIME_END/3600)) $(($TIME_END%3600/60)) $(($TIME_END%60))
	printf "\\tTo verify your installation run the following commands:\\n"
	printf "\\tFor more information:\\n"
	printf "\\tAnnChain website: www.annchain.io/#/\\n"
	printf "\\tAnnChain Telegram channel @ www.annchain.io/#/news\\n"
	printf "\\tAnnChain resources: https://github.com/dappledger/AnnChain\\n"
	printf "\\tAnnChain Stack Exchange: https://...\\n"
	printf "\\tAnnChain wiki: https://github.com/dappledger/AnnChain/blob/master/README.md\\n\\n\\n"
				
    else 
	echo "service not ok"
    fi
}

if ! [ -x "$(command -v git)" ];then
    InstallGit
else
    echo "Git has been installed!!!!"
fi

if ! [ -x "$(command -v go)" ];then
    InstallGo
else 
    echo "Golang has been installed !!!"
fi

BuildGenesis
InitGenesis
ConfigToml
RunGenesis
CheckGenesisIsOK
