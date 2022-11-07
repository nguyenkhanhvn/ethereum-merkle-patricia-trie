package nodes

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/utils"
)

type LeafNode struct {
	Path  []utils.Nibble
	Value []byte
}

func NewLeafNodeFromNibbleBytes(nibbles []byte, value []byte) (*LeafNode, error) {
	ns, err := utils.FromNibbleBytes(nibbles)
	if err != nil {
		return nil, fmt.Errorf("could not leaf node from nibbles: %w", err)
	}

	return NewLeafNodeFromNibbles(ns, value), nil
}

func NewLeafNodeFromNibbles(nibbles []utils.Nibble, value []byte) *LeafNode {
	return &LeafNode{
		Path:  nibbles,
		Value: value,
	}
}

func NewLeafNodeFromKeyValue(key, value string) *LeafNode {
	return NewLeafNodeFromBytes([]byte(key), []byte(value))
}

func NewLeafNodeFromBytes(key, value []byte) *LeafNode {
	return NewLeafNodeFromNibbles(utils.FromBytes(key), value)
}

func (l LeafNode) Hash() ([]byte, error) {
	serial, err := l.Serialize()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(serial), nil
}

func (l LeafNode) Raw() ([]interface{}, error) {
	path := utils.ToBytes(utils.ToPrefixed(l.Path, true))
	raw := []interface{}{path, l.Value}
	return raw, nil
}

func (l LeafNode) Serialize() ([]byte, error) {
	return Serialize(l)
}
