package nodes

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/utils"
)

type ExtensionNode struct {
	Path []utils.Nibble
	Next Node
}

func NewExtensionNode(nibbles []utils.Nibble, next Node) *ExtensionNode {
	return &ExtensionNode{
		Path: nibbles,
		Next: next,
	}
}

func (e ExtensionNode) Hash() ([]byte, error) {
	serial, err := e.Serialize()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(serial), nil
}

func (e ExtensionNode) Raw() ([]interface{}, error) {
	hashes := make([]interface{}, 2)
	hashes[0] = utils.ToBytes(utils.ToPrefixed(e.Path, false))
	if serial, err := Serialize(e.Next); err == nil && len(serial) >= 32 {
		hashes[1], err = e.Next.Hash()
		if err != nil {
			return nil, err
		}
	} else {
		hashes[1], err = e.Next.Raw()
		if err != nil {
			return nil, err
		}
	}
	return hashes, nil
}

func (e ExtensionNode) Serialize() ([]byte, error) {
	return Serialize(e)
}
