#!/bin/bash

inputDir=./cases
outputDir=./out
loader=./mapred
bencher=rzb

send() {
    action=$1
    outFile=$2
    procNum=$3
    requestNum=$4
    ip=$5

    test -z $ip && ip=localhost:46657

    # 1. create
    ./mapred --mapper="./rzb -action=${action} -num=${requestNum} ${ip}" --count=$procNum > $outFile
    start=`grep "START" $outFile | awk '{print $2}' | sort -n | head -1`
    end=`grep "END" $outFile | awk '{print $2}' | sort -n | tail -1`
    interval=`expr $end - $start`
    # num=`cat ${inputDir}/${caseName} | wc -l`
    ftps=`echo "${requestNum}*${procNum}/$interval" | bc -l`
    tps=`echo "$ftps*1000" | bc -l`
    echo "[ProcNum ]: $procNum"
    echo "[Requests]: $requestNum"
    echo "[Interval]: ${interval}ms"
    echo "[TPS     ]: $tps"

    # 2. read
    # ${binDir}/${loader} --mapper='${bencher} -action=read localhost:46657' --count=$procNum > $outFile_read

    # 3. call
    # ${binDir}/${loader} --mapper='${bencher} -action=call localhost:46657' --count=$procNum > $outFile_call
}

function main() {
    send $1 $2 $3 $4 $5
    #for req in `ls ${inputDir}`
    #do
    #    send ${req} ${procNum}
    #done

}

main $*