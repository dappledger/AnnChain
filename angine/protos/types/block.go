package types

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tmlibs/merkle"
)

func (blockID *BlockID) IsZero() bool {
	return len(blockID.Hash) == 0 && blockID.PartsHeader.IsZero()
}

func (blockID *BlockID) Equals(other *BlockID) bool {
	if blockID == other {
		return true
	}
	if blockID == nil || other == nil {
		return false
	}
	return bytes.Equal(blockID.Hash, other.Hash) &&
		blockID.PartsHeader.Equals(other.PartsHeader)
}

func (blockID *BlockID) Key() string {
	if blockID == nil {
		return ""
	}
	headerBys, _ := proto.Marshal(blockID.PartsHeader)
	return string(blockID.Hash) + string(headerBys)
}

func (blockID *BlockID) CString() string {
	if blockID == nil {
		return "nil"
	}
	return fmt.Sprintf(`%X:%v`, blockID.Hash, blockID.PartsHeader.CString())
}

/////////////////////////////////////////////////////////////////////////////

// NOTE: hash is nil if required fields are missing.
func (h *Header) Hash() []byte {
	if len(h.ValidatorsHash) == 0 {
		return nil
	}
	return merkle.SimpleHashFromMap(map[string]interface{}{
		"ChainID":            h.ChainID,
		"Height":             h.Height,
		"Time":               h.Time,
		"NumTxs":             h.NumTxs,
		"maker":              h.Maker,
		"LastBlockID":        h.LastBlockID,
		"LastCommit":         h.LastCommitHash,
		"Data":               h.DataHash,
		"Validators":         h.ValidatorsHash,
		"App":                h.AppHash,
		"Receipts":           h.ReceiptsHash,
		"LastNonEmptyHeight": h.LastNonEmptyHeight,
	})
}

func (h *Header) StringIndented(indent string) string {
	if h == nil {
		return "nil-Header"
	}
	return fmt.Sprintf(`Header{
%s  ChainID:        %v
%s  Height:         %v
%s  Time:           %v
%s  NumTxs:         %v
%s  Maker:          %X
%s  LastBlockID:    %v
%s  LastCommit:     %X
%s  Data:           %X
%s  Validators:     %X
%s  App:            %X
%s  Receipts:       %X
%s  LastNonEmptyHeight:       %v
%s}#%X`,
		indent, h.ChainID,
		indent, h.Height,
		indent, h.Time,
		indent, h.NumTxs,
		indent, h.Maker,
		indent, h.LastBlockID,
		indent, h.LastCommitHash,
		indent, h.DataHash,
		indent, h.ValidatorsHash,
		indent, h.AppHash,
		indent, h.ReceiptsHash,
		indent, h.LastNonEmptyHeight,
		indent, h.Hash(),
	)
}
