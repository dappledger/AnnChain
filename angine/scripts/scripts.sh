#! /bin/bash

PROTO_FILES=$(find ./protos/ -name *.proto)
PROTO_DIRS=$(dirname $PROTO_FILES | sort | uniq)

MSGPACK_FILES=("./types/events.go")

function buildProtos()
{
	for dir in $PROTO_DIRS; do
		protoc --proto_path=$GOPATH/src --proto_path=$dir --gofast_out=$dir $dir/*.proto
	done
	echo "build ok"
}

function clearProtos()
{
	for dir in $PROTO_DIRS; do
		rm -f $dir/*.pb.go
	done
	echo "clear done"
}

function buildMsgPack(){
	# go generate ./types/events.go
	for msgfile in $MSGPACK_FILES; do
		go generate $msgfile
	done
	echo "build ok"
}

function clearMsgPack(){
	for msgfile in $MSGPACK_FILES; do
		filename=${msgfile%.*}
		#echo filename:$filename
		rm -f ${filename}_gen.go ${filename}_gen_test.go
	done
	echo "clear done"
}

case $1 in
	buildpb)
		buildProtos
		;;
	clearpb)
		clearProtos
		;;
	buildmsgp)
		buildMsgPack
		;;
	clearmsgp)
		clearMsgPack
		;;
	all)
		clearProtos
		clearMsgPack
		buildProtos
		buildMsgPack
		;;
	*)
		echo missing params:buildpb/clearpb/buildmsgp/clearmsgp/...
		exit 1
		;;
esac

