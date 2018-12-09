#!/usr/bin/env bash

function wait_five_seconds {
    echo "Now wait for the creation of the organization"

    timer=0
    while ((timer < 5))
    do
        let "timer++"
        echo -en $timer "\r"
        sleep 1
    done
}

ANNPATH="/home/kyli/Codes/golang/src/github.com/dappledger/AnnChain/build"
CHAINID="annchain-gGgTjd"
ORGID="ikhofi1"

PRIV1="9F04A3EB2E3B412617F0A9D39466B357EBD3A073C28D004C73E482544515898D0FC4E216FB4B40781CEFAECB6C359BA6549069475B7DD678AECF1DF4AC5FCB4E"
PRIV2="2E28B2B44378E96048C777B1D48E5FB2AA3D88F295CC6D3371938C5B7820C8C47B039128642136EF7303A5CFAB5C719F705FCF1C08CA981DDEDCE1001932E12E"
PRIV3="CC1584C9ED0B56A5E1232E5F29425FE9E164EE01C5CA289F6017ED7143374867BC8DDD6A76E328F557069DBEA251D752F5EB284BF97583D2E37FD600783E83ED"
PRIV4="CBAF971D0B9CB4DD005A35878D5A31A5D9C27181EAE1091AC3300A80D6108076C8722AB9C75E00817463AF0BD2AD95B10A95BC3493B494B7016E1EC224D812D2"

echo "Creating ${ORGID} ..."

${ANNPATH}/anntool --callmode="commit" --backend="tcp://127.0.0.1:16657" --target="annchain-gGgTjd" organization create --genesisfile ./shards/ikhofi1/genesis1.json --configfile ./shards/ikhofi1/config1.json --privkey="${PRIV1}"

if [ 0 -ne $? ];then
    echo "Error: create org"
    exit 127
fi

wait_five_seconds

echo "Start joining..."

${ANNPATH}/anntool --callmode="commit" --backend="tcp://127.0.0.1:26657" --target="annchain-gGgTjd" organization join --configfile ./shards/ikhofi1/config2.json --privkey="${PRIV2}" --orgid="${ORGID}"

if [ 0 -ne $? ];then
    echo "Error: node 2"
    exit 127
fi

${ANNPATH}/anntool --callmode="commit" --backend="tcp://127.0.0.1:36657" --target="annchain-gGgTjd" organization join --configfile ./shards/ikhofi1/config3.json --privkey="${PRIV3}" --orgid="${ORGID}"

if [ 0 -ne $? ];then
    echo "Error: node 3"
    exit 127
fi

${ANNPATH}/anntool --callmode="commit" --backend="tcp://127.0.0.1:36657" --target="annchain-gGgTjd" organization join --configfile ./shards/ikhofi1/config4.json --privkey="${PRIV4}" --orgid="${ORGID}"

if [ 0 -ne $? ];then
    echo "Error: node 4"
    exit 127
fi
