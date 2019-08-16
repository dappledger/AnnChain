#!/bin/sh

tryGetInternalPkgs()
{
    unset GOPROXY
    # 解决jenkins与gitlab不通
    if [ ! -z "$HTTPS_PRXOY" ]
    then
        export https_proxy=$HTTPS_PRXOY
    fi
    go mod download
}

tryGetBlockedPkgs()
{
	export GOPROXY=https://goproxy.io
    if [ ! -z "$HTTPS_PRXOY" ]
    then
        unset https_proxy
    fi
    go mod download
}

# trick make
tryMake()
{
    t=`expr $1 - 1`;
    if [ $t == 0 ] ;then
        exit 1
    fi

	tryGetBlockedPkgs
	if [ $? -ne 0 ]; then
    	tryGetInternalPkgs
		if [ $? -ne 0 ]; then
		    tryMake $t
        fi
    fi
}

tryMake 10
