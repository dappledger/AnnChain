package trace

import (
	"errors"
	fmt "fmt"
	"reflect"

	proto "github.com/golang/protobuf/proto"
)

type TraceMsgItfc proto.Message

func UnmarshalTrcMsg(bz []byte) (TraceMsgItfc, error) {
	var trcMsg TraceMessage
	err := proto.Unmarshal(bz, &trcMsg)
	if err != nil {
		return nil, err
	}
	var msgItfc TraceMsgItfc
	switch trcMsg.GetType() {
	case MsgType_Request:
		msgItfc = &TraceRequest{}
	case MsgType_Responce:
		msgItfc = &TraceResponse{}
	default:
		return nil, errors.New(fmt.Sprintf("unmarshal,unknown consensus proto msg type:%v", reflect.TypeOf(msgItfc)))
	}
	err = proto.Unmarshal(trcMsg.GetData(), msgItfc)
	return msgItfc, err
}

func GetMessageType(msg proto.Message) MsgType {
	switch msg.(type) {
	case *TraceRequest:
		return MsgType_Request
	case *TraceResponse:
		return MsgType_Responce
	}
	return MsgType_None
}

func MarshalDataToTrcMsg(msg proto.Message) []byte {
	msgType := GetMessageType(msg)
	if msgType == MsgType_None {
		return nil
	}
	var bs, finbs []byte
	var err error
	bs, err = proto.Marshal(msg)
	if err != nil {
		return nil
	}
	msgBase := &TraceMessage{
		Type: msgType,
		Data: bs,
	}
	finbs, err = proto.Marshal(msgBase)
	if err != nil {
		return nil
	}
	return finbs

}
