package consensus

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
)

type StConsensusMsg struct {
	ConsensusMsgItfc
}

type cssMsgJson struct {
	MsgType MsgType `json:"msg_type"`
	JsonBys []byte  `json:"json_str"`
}

func (csm StConsensusMsg) MarshalJSON() ([]byte, error) {
	st := &cssMsgJson{}
	st.MsgType = GetMessageType(csm.ConsensusMsgItfc)
	jbys, err := json.Marshal(&csm.ConsensusMsgItfc)
	if err != nil {
		return nil, err
	}
	st.JsonBys = jbys
	return json.Marshal(&st)
}

func (csm *StConsensusMsg) UnmarshalJSON(data []byte) error {
	var dec cssMsgJson
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	csm.ConsensusMsgItfc, err = unmarshalJson(dec.MsgType, dec.JsonBys)
	return err
}

type ConsensusMsgItfc interface {
	proto.Message
	CString() string
}

func unmarshalJson(typ MsgType, jsonBys []byte) (ConsensusMsgItfc, error) {
	var err error
	switch typ {
	case MsgType_NewRoundStep:
		data := &NewRoundStepMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_CommitStep:
		data := &CommitStepMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_Proposal:
		data := &ProposalMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_ProposalPOL:
		data := &ProposalPOLMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_BlockPart:
		data := &BlockPartMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_Vote:
		data := &VoteMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_HasVote:
		data := &HasVoteMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_VoteSetMaj23:
		data := &VoteSetMaj23Message{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	case MsgType_VoteSetBits:
		data := &VoteSetBitsMessage{}
		err = json.Unmarshal(jsonBys, &data)
		return data, err
	default:
		return nil, errors.New(fmt.Sprintf("unmarshal,unknown consensus proto msg type:%v", typ))
	}
}

func MsgFromType(typ MsgType) (msgItfc ConsensusMsgItfc, err error) {
	switch typ {
	case MsgType_NewRoundStep:
		msgItfc = &NewRoundStepMessage{}
	case MsgType_CommitStep:
		msgItfc = &CommitStepMessage{}
	case MsgType_Proposal:
		msgItfc = &ProposalMessage{}
	case MsgType_ProposalPOL:
		msgItfc = &ProposalPOLMessage{}
	case MsgType_BlockPart:
		msgItfc = &BlockPartMessage{}
	case MsgType_Vote:
		msgItfc = &VoteMessage{}
	case MsgType_HasVote:
		msgItfc = &HasVoteMessage{}
	case MsgType_VoteSetMaj23:
		msgItfc = &VoteSetMaj23Message{}
	case MsgType_VoteSetBits:
		msgItfc = &VoteSetBitsMessage{}
	default:
		return nil, errors.New(fmt.Sprintf("unmarshal,unknown consensus msg type:%v", typ))
	}
	return
}

func UnmarshalCssMsg(bz []byte) (ConsensusMsgItfc, error) {
	var cssMsg ConsensusMessage
	err := proto.Unmarshal(bz, &cssMsg)
	if err != nil {
		return nil, err
	}
	var msgItfc ConsensusMsgItfc
	msgItfc, err = MsgFromType(cssMsg.GetType())
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(cssMsg.GetData(), msgItfc)
	return msgItfc, err
}

func GetMessageType(msg proto.Message) MsgType {
	switch msg.(type) {
	case *NewRoundStepMessage:
		return MsgType_NewRoundStep
	case *CommitStepMessage:
		return MsgType_CommitStep
	case *ProposalMessage:
		return MsgType_Proposal
	case *ProposalPOLMessage:
		return MsgType_ProposalPOL
	case *BlockPartMessage:
		return MsgType_BlockPart
	case *VoteMessage:
		return MsgType_Vote
	case *HasVoteMessage:
		return MsgType_HasVote
	case *VoteSetMaj23Message:
		return MsgType_VoteSetMaj23
	case *VoteSetBitsMessage:
		return MsgType_VoteSetBits
	default:
		//PanicCrisis(fmt.Sprintf("gettype,unknown consensus proto msg type:%v", reflect.TypeOf(msg)))
	}
	return MsgType_None
}

func MarshalDataToCssMsg(msg proto.Message) []byte {
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
	msgBase := &ConsensusMessage{
		Type: msgType,
		Data: bs,
	}
	finbs, err = proto.Marshal(msgBase)
	if err != nil {
		return nil
	}
	return finbs
}

// Must ok
func MarshalData(msg ConsensusMsgItfc) []byte {
	bs, _ := proto.Marshal(msg)
	return bs
}

func TransferBitArray(bitArray *BitArray) (p *ProtoBitArray) {
	if bitArray == nil {
		return nil
	}
	p = &ProtoBitArray{
		Bits:  int64(bitArray.Bits),
		Elems: bitArray.Elems,
	}
	return
}

func TransferProtoBitArray(p *ProtoBitArray) (ba *BitArray) {
	if p == nil {
		return nil
	}
	ba = &BitArray{
		Bits:  int(p.Bits),
		Elems: p.Elems,
	}
	return
}

//-------------------------------------

// For every height/round/step transition
func (m *NewRoundStepMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[NewRoundStep H:%v R:%v S:%v LCR:%v]",
		m.Height, m.Round, m.Step, m.LastCommitRound)
}

//-------------------------------------

//BlockParts       *BitArray

func (m *CommitStepMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[CommitStep H:%v BP:%v BA:%v]", m.Height, m.BlockPartsHeader, m.BlockParts)
}

//-------------------------------------

func (m *ProposalMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[Proposal %v]", m.Proposal.CString())
}

//-------------------------------------

//ProposalPOL      *BitArray

func (m *ProposalPOLMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[ProposalPOL H:%v POLR:%v POL:%v]", m.Height, m.ProposalPOLRound, m.ProposalPOL)
}

//-------------------------------------

func (m *BlockPartMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[BlockPart H:%v R:%v P:%v]", m.Height, m.Round, m.Part)
}

//-------------------------------------

func (m *VoteMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[Vote %v]", m.Vote)
}

//-------------------------------------

func (m *HasVoteMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[HasVote VI:%v V:{%v/%02d/%v} VI:%v]", m.Index, m.Height, m.Round, m.Type, m.Index)
}

//-------------------------------------

func (m *VoteSetMaj23Message) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[VSM23 %v/%02d/%v %v]", m.Height, m.Round, m.Type, m.BlockID)
}

//-------------------------------------

//Votes   *BitArray

func (m *VoteSetBitsMessage) CString() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("[VSB %v/%02d/%v %v %v]", m.Height, m.Round, m.Type, m.BlockID, m.Votes)
}

func (rs RoundStepType) CString() string {
	switch rs {
	case RoundStepType_NewHeight:
		return "RoundStepNewHeight"
	case RoundStepType_NewRound:
		return "RoundStepNewRound"
	case RoundStepType_Propose:
		return "RoundStepPropose"
	case RoundStepType_Prevote:
		return "RoundStepPrevote"
	case RoundStepType_PrevoteWait:
		return "RoundStepPrevoteWait"
	case RoundStepType_Precommit:
		return "RoundStepPrecommit"
	case RoundStepType_PrecommitWait:
		return "RoundStepPrecommitWait"
	case RoundStepType_Commit:
		return "RoundStepCommit"
	default:
		return "RoundStepUnknown" // Cannot panic.
	}
}
