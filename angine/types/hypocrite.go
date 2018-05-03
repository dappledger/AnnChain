package types

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	"github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

const (
	_ byte = iota
	REASON_CONNECTION
	REASON_BADVOTE
	REASON_PROPOSETO
)

type (
	IBadVoteCollector interface {
		ReportBadVote(crypto.PubKey, interface{})
	}

	IDisconnectCollector interface {
		ReportDisconnect(crypto.PubKey, interface{})
	}

	Hypocrite struct {
		PubKey   string
		Reason   byte
		Evidence interface{}
		Time     time.Time
	}

	Hypocrites []*Hypocrite

	HypoDisconnectEvidence struct {
		PubKey string
	}

	HypoBadVoteEvidence struct {
		PubKey   string
		VoteType pbtypes.VoteType
		Height   def.INT
		Round    def.INT
		Got      *pbtypes.Vote
		Expected *pbtypes.Vote
	}

	HypoProposalTimeoutEvidence struct {
		Proposal     *pbtypes.Proposal
		BlockPartsBA *common.BitArray
		Height       def.INT
		Round        def.INT
	}

	iHypo struct {
		PubKey   string    `json:"pubkey"`
		Reason   byte      `json:"reason"`
		Evidence []byte    `json:"evidence"`
		Time     time.Time `json:"time"`
	}
)

func NewHypocrite(pk string, reason byte, evidence interface{}) *Hypocrite {
	return &Hypocrite{
		PubKey:   pk,
		Reason:   reason,
		Evidence: evidence,
		Time:     time.Now(),
	}
}

func NewHypocrites() Hypocrites {
	return make([]*Hypocrite, 0)
}

func (h *Hypocrite) MarshalJSON() ([]byte, error) {
	ih := iHypo{
		PubKey: h.PubKey,
		Reason: h.Reason,
		Time:   h.Time,
	}

	evbytes, err := json.Marshal(h.Evidence)
	if err != nil {
		return nil, errors.Wrap(err, "[Hypocrite MarshalJSON]")
	}

	ih.Evidence = evbytes

	return json.Marshal(ih)
}

func (h *Hypocrite) UnmarshalJSON(bs []byte) error {
	ih := &iHypo{}
	if err := json.Unmarshal(bs, ih); err != nil {
		return errors.Wrap(err, "[SuspectTx UnmarshalJSON]")
	}

	h.Time = ih.Time
	h.Reason = ih.Reason
	h.PubKey = ih.PubKey

	switch ih.Reason {
	case REASON_BADVOTE:
		ev := &HypoBadVoteEvidence{}
		if err := json.Unmarshal(ih.Evidence, ev); err != nil {
			return errors.Wrap(err, "[Hypocrite UnmarshalJSON]")
		}
		h.Evidence = ev
	case REASON_CONNECTION:

	case REASON_PROPOSETO:
		ev := &HypoProposalTimeoutEvidence{}
		if err := json.Unmarshal(ih.Evidence, ev); err != nil {
			return errors.Wrap(err, "[Hypocrite UnmarshalJSON]")
		}
		h.Evidence = ev
	default:

	}

	return nil
}
