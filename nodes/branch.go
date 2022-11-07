package nodes

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/utils"
)

type BranchNode struct {
	Branches [16]Node
	Value    []byte
}

func NewBranchNode() *BranchNode {
	return &BranchNode{
		Branches: [16]Node{},
	}
}

func (b BranchNode) Hash() ([]byte, error) {
	serial, err := b.Serialize()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(serial), nil
}

func (b *BranchNode) SetBranch(nibble utils.Nibble, node Node) {
	b.Branches[int(nibble)] = node
}

func (b *BranchNode) RemoveBranch(nibble utils.Nibble) {
	b.Branches[int(nibble)] = nil
}

func (b *BranchNode) SetValue(value []byte) {
	b.Value = value
}

func (b *BranchNode) RemoveValue() {
	b.Value = nil
}

func (b BranchNode) Raw() ([]interface{}, error) {
	hashes := make([]interface{}, 17)
	for i := 0; i < 16; i++ {
		if b.Branches[i] == nil {
			hashes[i] = EmptyNodeRaw
		} else {
			node := b.Branches[i]
			if serial, err := Serialize(node); err == nil && len(serial) >= 32 {
				hashes[i], err = node.Hash()
				if err != nil {
					return nil, err
				}
			} else {
				// if node can be serialized to less than 32 bits, then
				// use Serialized directly.
				// it has to be ">=", rather than ">",
				// so that when deserialized, the content can be distinguished
				// by length
				hashes[i], err = node.Raw()
				if err != nil {
					return nil, err
				}
			}
		}
	}

	hashes[16] = b.Value
	return hashes, nil
}

func (b BranchNode) Serialize() ([]byte, error) {
	return Serialize(b)
}

func (b BranchNode) HasValue() bool {
	return b.Value != nil
}
