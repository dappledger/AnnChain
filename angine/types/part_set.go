// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/ripemd160"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-merkle"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

var (
	ErrPartSetUnexpectedIndex = errors.New("Error part set unexpected index")
	ErrPartSetInvalidProof    = errors.New("Error part set invalid proof")
)

type PartCache struct {
	*pbtypes.Part
	hash []byte
}

func NewPartCache(p *pbtypes.Part) *PartCache {
	return &PartCache{
		Part: p,
	}
}

func (part *PartCache) init(aunt [][]byte) {
	if part == nil {
		part = &PartCache{}
	}
	if part.Part == nil {
		part.Part = &pbtypes.Part{}
	}
	if part.Part.Proof == nil {
		part.Part.Proof = &pbtypes.SimpleProof{}
	}
	part.Part.Proof.Bytes = aunt
}

func (part *PartCache) Hash() []byte {
	if part.hash != nil {
		return part.hash
	}
	hasher := ripemd160.New()
	hasher.Write(part.Bytes) // doesn't err
	part.hash = hasher.Sum(nil)
	return part.hash
}

func (part *PartCache) CString() string {
	if part == nil {
		return "nil"
	}
	return part.StringIndented("")
}

func (part *PartCache) StringIndented(indent string) string {
	proof := part.MerkleProof()
	return fmt.Sprintf(`Part{#%v
%s  Bytes: %X...
%s  Proof: %v
%s}`,
		part.Index,
		indent, Fingerprint(part.Bytes),
		indent, (&proof).StringIndented(indent+"  "),
		indent)
}

//-------------------------------------

type PartSet struct {
	total def.INT
	hash  []byte

	mtx           sync.Mutex
	parts         []*PartCache
	partsBitArray *BitArray
	count         def.INT
}

// Returns an immutable, full PartSet from the data bytes.
// The data bytes are split into "partSize" chunks, and merkle tree computed.
func NewPartSetFromData(data []byte, partSize def.INT) *PartSet {
	// divide data into 4kb parts.
	intPartSize := int(partSize)
	total := (len(data) + intPartSize - 1) / intPartSize
	parts := make([]*PartCache, total)
	parts_ := make([]merkle.Hashable, total)
	partsBitArray := NewBitArray(total)
	for i := 0; i < total; i++ {
		part := &PartCache{
			Part: &pbtypes.Part{
				Index: def.HINT(i),
				Bytes: data[i*intPartSize : MinInt(len(data), (i+1)*intPartSize)],
			},
		}
		parts[i] = part
		parts_[i] = part
		partsBitArray.SetIndex(i, true)
	}
	// Compute merkle proofs
	root, proofs := merkle.SimpleProofsFromHashables(parts_)
	for i := 0; i < total; i++ {
		parts[i].init(proofs[i].Aunts)
	}
	return &PartSet{
		total:         def.INT(total),
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         def.INT(total),
	}
}

// Returns an empty PartSet ready to be populated.
func NewPartSetFromHeader(header *pbtypes.PartSetHeader) *PartSet {
	return &PartSet{
		total:         def.INT(header.Total),
		hash:          header.Hash,
		parts:         make([]*PartCache, int(header.Total)),
		partsBitArray: NewBitArray(int(header.Total)),
		count:         0,
	}
}

func (ps *PartSet) Header() *pbtypes.PartSetHeader {
	if ps == nil {
		return &pbtypes.PartSetHeader{}
	}
	return &pbtypes.PartSetHeader{
		Total: def.HINT(ps.total),
		Hash:  ps.hash,
	}
}

func (ps *PartSet) HasHeader(header *pbtypes.PartSetHeader) bool {
	if ps == nil {
		return false
	}
	myheader := ps.Header()
	return myheader.Equals(header)
}

func (ps *PartSet) BitArray() *BitArray {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.partsBitArray.Copy()
}

func (ps *PartSet) Hash() []byte {
	if ps == nil {
		return nil
	}
	return ps.hash
}

func (ps *PartSet) HashesTo(hash []byte) bool {
	if ps == nil {
		return false
	}
	return bytes.Equal(ps.hash, hash)
}

func (ps *PartSet) Count() def.INT {
	if ps == nil {
		return 0
	}
	return ps.count
}

func (ps *PartSet) Total() def.INT {
	if ps == nil {
		return 0
	}
	return ps.total
}

func (ps *PartSet) AddPart(part *pbtypes.Part, verify bool) (bool, error) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	intIndex := int(part.Index)
	// Invalid part index
	if def.INT(intIndex) >= ps.total {
		return false, ErrPartSetUnexpectedIndex
	}

	// If part already exists, return false.
	if ps.parts[intIndex] != nil {
		return false, nil
	}

	// Check hash proof
	partcache := &PartCache{Part: part}
	if verify {
		proof := part.MerkleProof()
		if !(&proof).Verify(intIndex, int(ps.total), partcache.Hash(), ps.Hash()) {
			return false, ErrPartSetInvalidProof
		}
	}

	// Add part
	ps.parts[intIndex] = partcache
	ps.partsBitArray.SetIndex(intIndex, true)
	ps.count++
	return true, nil
}

func (ps *PartSet) GetPart(index int) *pbtypes.Part {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parts[index].Part
}

func (ps *PartSet) IsComplete() bool {
	return ps.count == ps.total
}

func (ps *PartSet) StringShort() string {
	if ps == nil {
		return "nil-PartSet"
	}
	return fmt.Sprintf("(%v of %v)", ps.Count(), ps.Total())
}

func (ps *PartSet) AssembleToBlock(partSize def.INT) *BlockCache {
	if len(ps.parts) == 0 {
		return nil
	}
	intPartSize := int(partSize)
	bysLen := intPartSize*(len(ps.parts)-1) + len(ps.parts[len(ps.parts)-1].Part.GetBytes())
	//bys := make([]byte, 0, bysLen)
	buff := bytes.Buffer{}
	buff.Grow(bysLen)
	for i := range ps.parts {
		//bys = append(bys, ps.parts[i].Part.GetBytes()...)
		buff.Write(ps.parts[i].Part.GetBytes())
	}
	var block pbtypes.Block
	UnmarshalData(buff.Bytes(), &block)
	return MakeBlockCache(&block)
}
