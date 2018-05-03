package tools

import (
	"github.com/gogo/protobuf/proto"
)

func PbMarshal(msg proto.Message) []byte {
	ret, err := proto.Marshal(msg)
	if err != nil {
		return nil
	}
	return ret
}

func PbUnmarshal(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

func CopyBytes(byts []byte) []byte {
	cp := make([]byte, len(byts))
	copy(cp, byts)
	return cp
}
