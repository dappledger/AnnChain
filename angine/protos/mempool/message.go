package mempool

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

//-----------------------------------------------------------------------------
// Messages

type MempoolMsgItfc proto.Message

func UnmarshalMpMsg(bz []byte) (MempoolMsgItfc, error) {
	var mpMsg MempoolMessage
	err := proto.Unmarshal(bz, &mpMsg)
	if err != nil {
		return nil, err
	}
	var msgItfc MempoolMsgItfc
	switch mpMsg.GetType() {
	case MsgType_Tx:
		msgItfc = &TxMessage{}
	}
	err = proto.Unmarshal(mpMsg.GetData(), msgItfc)
	return msgItfc, err
}

func GetMessageType(msg proto.Message) MsgType {
	switch msg.(type) {
	case *TxMessage:
		return MsgType_Tx
	}
	return MsgType_None
}

func MarshalDataToMpMsg(msg proto.Message) []byte {
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
	msgBase := &MempoolMessage{
		Type: msgType,
		Data: bs,
	}
	finbs, err = proto.Marshal(msgBase)
	if err != nil {
		return nil
	}
	return finbs
}

func (m *TxMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[TxMessage %v]", m.Tx)
}
