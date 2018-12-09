/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package iavl

import (
	"bytes"

	"golang.org/x/crypto/ripemd160"

	"github.com/golang/protobuf/proto"
	"github.com/tendermint/go-wire"
	. "github.com/tendermint/tmlibs/common"
	pb "github.com/dappledger/AnnChain/module/xlib/iavl/pb"
)

const proofLimit = 1 << 16 // 64 KB

type IAVLProof struct {
	LeafHash   []byte
	InnerNodes []IAVLProofInnerNode
	RootHash   []byte
}

func (proof *IAVLProof) Verify(key []byte, value []byte, root []byte) bool {
	if !bytes.Equal(proof.RootHash, root) {
		return false
	}
	leafNode := IAVLProofLeafNode{KeyBytes: key, ValueBytes: value}
	leafHash := leafNode.Hash()
	if !bytes.Equal(leafHash, proof.LeafHash) {
		return false
	}
	hash := leafHash
	for _, branch := range proof.InnerNodes {
		hash = branch.Hash(hash)
	}
	return bytes.Equal(proof.RootHash, hash)
}

// Please leave this here!  I use it in light-client to fulfill an interface
func (proof *IAVLProof) Root() []byte {
	return proof.RootHash
}

// ReadProof will deserialize a IAVLProof from bytes
func ReadProof(data []byte) (*IAVLProof, error) {
	// TODO: make go-wire never panic
	n, err := int(0), error(nil)
	proof := wire.ReadBinary(&IAVLProof{}, bytes.NewBuffer(data), proofLimit, &n, &err).(*IAVLProof)
	return proof, err

}

type IAVLProofInnerNode struct {
	Height int8
	Size   int
	Left   []byte
	Right  []byte
}

func (branch IAVLProofInnerNode) Hash(childHash []byte) []byte {
	hasher := ripemd160.New()
	//buf := new(bytes.Buffer)
	//n, err := int(0), error(nil)
	//
	//
	//
	//wire.WriteInt8(branch.Height, buf, &n, &err)
	//wire.WriteVarint(branch.Size, buf, &n, &err)
	//if len(branch.Left) == 0 {
	//	wire.WriteByteSlice(childHash, buf, &n, &err)
	//	wire.WriteByteSlice(branch.Right, buf, &n, &err)
	//} else {
	//	wire.WriteByteSlice(branch.Left, buf, &n, &err)
	//	wire.WriteByteSlice(childHash, buf, &n, &err)
	//}

	pbpin := &pb.IAVLProofInnerNode{}

	pbpin.Height = uint32(branch.Height)
	pbpin.Size = uint64(branch.Size)
	if len(branch.Left) == 0 {
		pbpin.Left = childHash
		pbpin.Right = branch.Right
	} else {
		pbpin.Left = branch.Left
		pbpin.Right = branch.Right
	}

	buff, err := proto.Marshal(pbpin)

	if err != nil {
		PanicCrisis(Fmt("Failed to hash IAVLProofInnerNode: %v", err))
	}
	// fmt.Printf("InnerNode hash bytes: %X\n", buf.Bytes())

	//hasher.Write(buf.Bytes())

	hasher.Write(buff)
	return hasher.Sum(nil)
}

type IAVLProofLeafNode struct {
	KeyBytes   []byte
	ValueBytes []byte
}

func (leaf IAVLProofLeafNode) Hash() []byte {
	hasher := ripemd160.New()
	//buf := new(bytes.Buffer)
	//n, err := int(0), error(nil)
	//wire.WriteInt8(0, buf, &n, &err)
	//wire.WriteVarint(1, buf, &n, &err)
	//wire.WriteByteSlice(leaf.KeyBytes, buf, &n, &err)
	//wire.WriteByteSlice(leaf.ValueBytes, buf, &n, &err)
	//
	//
	//
	//if err != nil {
	//	PanicCrisis(Fmt("Failed to hash IAVLProofLeafNode: %v", err))
	//}
	//// fmt.Printf("LeafNode hash bytes:   %X\n", buf.Bytes())
	//hasher.Write(buf.Bytes())

	pbn := &pb.IAVLNode{}
	pbn.Height = 0
	pbn.Size = 1
	pbn.Key = leaf.KeyBytes
	pbn.Value = leaf.ValueBytes

	buff, err := proto.Marshal(pbn)

	if err != nil {
		PanicCrisis(Fmt("Failed to hash IAVLProofLeafNode: %v", err))
	}

	hasher.Write(buff)

	return hasher.Sum(nil)
}

func (node *IAVLNode) constructProof(t *IAVLTree, key []byte, valuePtr *[]byte, proof *IAVLProof) (exists bool) {
	if node.height == 0 {
		if bytes.Compare(node.key, key) == 0 {
			*valuePtr = node.value
			proof.LeafHash = node.hash
			return true
		} else {
			return false
		}
	} else {
		if bytes.Compare(key, node.key) < 0 {
			exists := node.getLeftNode(t).constructProof(t, key, valuePtr, proof)
			if !exists {
				return false
			}
			branch := IAVLProofInnerNode{
				Height: node.height,
				Size:   node.size,
				Left:   nil,
				Right:  node.getRightNode(t).hash,
			}
			proof.InnerNodes = append(proof.InnerNodes, branch)
			return true
		} else {
			exists := node.getRightNode(t).constructProof(t, key, valuePtr, proof)
			if !exists {
				return false
			}
			branch := IAVLProofInnerNode{
				Height: node.height,
				Size:   node.size,
				Left:   node.getLeftNode(t).hash,
				Right:  nil,
			}
			proof.InnerNodes = append(proof.InnerNodes, branch)
			return true
		}
	}
}

// Returns nil, nil if key is not in tree.
func (t *IAVLTree) ConstructProof(key []byte) (value []byte, proof *IAVLProof) {
	if t.root == nil {
		return nil, nil
	}
	t.root.hashWithCount(t) // Ensure that all hashes are calculated.
	proof = &IAVLProof{
		RootHash: t.root.hash,
	}
	exists := t.root.constructProof(t, key, &value, proof)
	if exists {
		return value, proof
	} else {
		return nil, nil
	}
}
