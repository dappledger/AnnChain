#!/bin/sh

base=`pwd`
bindir=$base/bin
prev=github.com/dappledger/AnnChain

begin()
{
	mkdir -p bin
    cd $base/../../../../
    export GOPATH=`pwd`
}

end()
{
	cd $base
}

build()
{
	echo $GOPATH
	echo "go build -o $bindir/$2 $1"
    go build -o $bindir/$2 $1
}

run()
{
    begin
    case $1 in
        genesis )
            build $prev"/cmd/genesis" "genesis"
            ;;
        gtool )
            build $prev"/cmd/client" "gtool"
            ;;
        * )
            build $prev"/cmd/genesis" "genesis"
            build $prev"/cmd/client" "gtool"
            ;;
    esac
	end
}

run $1
