package blockchain

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

type BlockMsgItfc proto.Message

func UnmarshalBlkMsg(bz []byte) (BlockMsgItfc, error) {
	var blkMsg BlockMessage
	err := proto.Unmarshal(bz, &blkMsg)
	if err != nil {
		return nil, err
	}
	var msgItfc BlockMsgItfc
	switch blkMsg.GetType() {
	case MsgType_BlockReq:
		msgItfc = &BlockRequestMessage{}
	case MsgType_BlockRsp:
		msgItfc = &BlockResponseMessage{}
	case MsgType_StatusReq:
		msgItfc = &StatusRequestMessage{}
	case MsgType_StatusRsp:
		msgItfc = &StatusResponseMessage{}
	case MsgType_HeaderReq:
		msgItfc = &BlockHeaderRequestMessage{}
	case MsgType_HeaderRsp:
		msgItfc = &BlockHeaderResponseMessage{}
	default:
		return nil, errors.New(fmt.Sprintf("unmarshal,unknown consensus proto msg type:%v", reflect.TypeOf(msgItfc)))
	}
	err = proto.Unmarshal(blkMsg.GetData(), msgItfc)
	return msgItfc, err
}

func GetMessageType(msg proto.Message) MsgType {
	switch msg.(type) {
	case *BlockRequestMessage:
		return MsgType_BlockReq
	case *BlockResponseMessage:
		return MsgType_BlockRsp
	case *StatusRequestMessage:
		return MsgType_StatusReq
	case *StatusResponseMessage:
		return MsgType_StatusRsp
	case *BlockHeaderRequestMessage:
		return MsgType_HeaderReq
	case *BlockHeaderResponseMessage:
		return MsgType_HeaderRsp
	default:
	}
	return MsgType_None
}

func MarshalDataToBlkMsg(msg proto.Message) []byte {
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
	msgBase := &BlockMessage{
		Type: msgType,
		Data: bs,
	}
	finbs, err = proto.Marshal(msgBase)
	if err != nil {
		return nil
	}
	return finbs
}

func MarshalData(msg BlockMsgItfc) []byte {
	bs, _ := proto.Marshal(msg)
	return bs
}

////////////////////////////////////////////////////////////////////////////////////

func (m *BlockRequestMessage) CString() string {
	if m == nil {
		return "nil"
	}
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[bcBlockRequestMessage %v]", m.Height)
}

func (m *BlockResponseMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[bcBlockResponseMessage %v]", m.Block.Header.Height)
}

func (m *StatusRequestMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[bcStatusRequestMessage %v]", m.Height)
}

func (m *StatusResponseMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[bcStatusResponseMessage %v]", m.Height)
}

func (m *BlockHeaderRequestMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[bcBlockHeaderRequestMessage %v]", m.Height)
}

func (m *BlockHeaderResponseMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[bcBlockHeaderResponseMessage %v]", m.Header.Height)
}
