#!/bin/sh
# chkconfig: 2345 64 36
# description: AnnChain.genesis startup
# Author:lvguoxin
# Time:2018-10-30 15:39:34
# Name:annchainService.sh
# Version:V1.0
# Description:This is a test script.

[ -f /etc/init.d/functions ] && source /etc/init.d/functions
GENESIS_EX_PATH=${GOPATH}/src/gitlab.zhonganinfo.com/tech_bighealth/za-delos/build
RETURN_VALUE=0


#log failure output func
function LogFailureMsg()
{
    echo "Genesis service ERROR!$@"
}

#log success output func
function LogSuccessMsg()
{
    echo "Genesis service SUCCESS!$@"
}

#Genesis start service
function start()
{
    cd ${GENESIS_EX_PATH}
    echo "start Genesis service"
    nohup ./genesis node >/dev/null 2>&1  &
    if [ $? -ne 0 ];then
        LogFailureMsg
    else
        LogSuccessMsg
    fi
}

#Genesis stop service
function stop()
{   
    echo "Stop Genesis service"
    kill `ps -ef | grep genesis | grep -v grep|awk '{print $2}'`  >/dev/null 2>&1 &
    if [ $? -ne 0 ];then
        LogFailureMsg
    else
        LogSuccessMsg
    fi
}


case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    *)
	echo "Usage:$0{start|stop|restart}"
        exit 1
esac
exit $RETURN_VALUE

