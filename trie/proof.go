package trie

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/nodes"
	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/utils"
)

// CreateProof including proofsibling: all node to key, path: key nibble, error
func (t *Trie) CreateProof(key []byte) ([][]byte, []byte, error) {
	var proof [][]byte
	var serial []byte
	var err error
	node := t.root
	nibbles := utils.FromBytes(key)
	keyNibble := nibbles
	for {
		if nodes.IsEmptyNode(node) {
			return nil, nil, fmt.Errorf("empty node")
		}

		if leaf, ok := node.(*nodes.LeafNode); ok {
			matched := utils.PrefixMatchedLen(leaf.Path, nibbles)
			if matched != len(leaf.Path) || matched != len(nibbles) {
				return nil, nil, fmt.Errorf("key not found")
			}
			serial, err = leaf.Serialize()
			if err != nil {
				return nil, nil, err
			}
			proof = append(proof, serial)

			keyNibble = utils.ToPrefixed(keyNibble, true)
			return proof, utils.ToBytes(keyNibble), nil
		}

		if branch, ok := node.(*nodes.BranchNode); ok {
			if len(nibbles) == 0 {
				serial, err = branch.Serialize()
				if err != nil {
					return nil, nil, err
				}
				proof = append(proof, serial)

				keyNibble := utils.ToPrefixed(keyNibble, false)
				if !branch.HasValue() {
					return proof, utils.ToBytes(keyNibble), fmt.Errorf("node has no value")
				}
				return proof, utils.ToBytes(keyNibble), nil
			}

			b, remaining := nibbles[0], nibbles[1:]
			nibbles = remaining
			node = branch.Branches[b]

			serial, err = branch.Serialize()
			if err != nil {
				return nil, nil, err
			}
			proof = append(proof, serial)
			continue
		}

		if ext, ok := node.(*nodes.ExtensionNode); ok {
			matched := utils.PrefixMatchedLen(ext.Path, nibbles)
			// E 01020304
			//   010203
			if matched < len(ext.Path) {
				return nil, nil, fmt.Errorf("key not found")
			}

			nibbles = nibbles[matched:]
			node = ext.Next

			serial, err = ext.Serialize()
			if err != nil {
				return nil, nil, err
			}
			proof = append(proof, serial)
			continue
		}

		return nil, nil, fmt.Errorf("key not found")
	}
}

type Proof interface {
	// Put inserts the given value into the key-value data store.
	Put(key []byte, value []byte) error

	// Delete removes the key from the key-value data store.
	Delete(key []byte) error

	// Has retrieves if a key is present in the key-value data store.
	Has(key []byte) (bool, error)

	// Get retrieves the given key if it's present in the key-value data store.
	Get(key []byte) ([]byte, error)

	// Serialize returns the serialized proof
	Serialize() [][]byte
}

type ProofDB struct {
	kv map[string][]byte
}

func NewProofDB() *ProofDB {
	return &ProofDB{
		kv: make(map[string][]byte),
	}
}

func (w *ProofDB) Put(key []byte, value []byte) error {
	keyS := hex.EncodeToString(key)
	w.kv[keyS] = value
	fmt.Printf("put key: %x, value: %x\n", keyS, value)
	return nil
}

func (w *ProofDB) Delete(key []byte) error {
	keyS := fmt.Sprintf("%x", key)
	delete(w.kv, keyS)
	return nil
}
func (w *ProofDB) Has(key []byte) (bool, error) {
	keyS := fmt.Sprintf("%x", key)
	_, ok := w.kv[keyS]
	return ok, nil
}

func (w *ProofDB) Get(key []byte) ([]byte, error) {
	keyS := fmt.Sprintf("%x", key)
	val, ok := w.kv[keyS]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return val, nil
}

func (w *ProofDB) Serialize() [][]byte {
	nodes := make([][]byte, 0, len(w.kv))
	for _, value := range w.kv {
		nodes = append(nodes, value)
	}
	return nodes
}

// Prove returns the merkle proof for the given key, which is
func (t *Trie) Prove(key []byte) (Proof, error) {
	proof := NewProofDB()
	node := t.root
	nibbles := utils.FromBytes(key)

	for {
		if nodes.IsEmptyNode(node) {
			return nil, fmt.Errorf("empty node")
		}

		hash, err := nodes.Hash(node)
		if err != nil {
			return nil, err
		}
		serial, err := nodes.Serialize(node)
		if err != nil {
			return nil, err
		}
		proof.Put(hash, serial)

		if leaf, ok := node.(*nodes.LeafNode); ok {
			matched := utils.PrefixMatchedLen(leaf.Path, nibbles)
			if matched != len(leaf.Path) || matched != len(nibbles) {
				return nil, fmt.Errorf("key not found")
			}

			return proof, nil
		}

		if branch, ok := node.(*nodes.BranchNode); ok {
			if len(nibbles) == 0 {
				if branch.HasValue() {
					return proof, nil
				} else {
					return proof, fmt.Errorf("node has no value")
				}
			}

			b, remaining := nibbles[0], nibbles[1:]
			nibbles = remaining
			node = branch.Branches[b]
			continue
		}

		if ext, ok := node.(*nodes.ExtensionNode); ok {
			matched := utils.PrefixMatchedLen(ext.Path, nibbles)
			// E 01020304
			//   010203
			if matched < len(ext.Path) {
				return nil, fmt.Errorf("key not found")
			}

			nibbles = nibbles[matched:]
			node = ext.Next
			continue
		}

		return nil, fmt.Errorf("key not found")
	}
}

// VerifyProof verify the proof for the given key under the given root hash using go-ethereum's VerifyProof implementation.
// It returns the value for the key if the proof is valid, otherwise error will be returned
func VerifyProof(rootHash []byte, key []byte, proof Proof) (value []byte, err error) {
	return trie.VerifyProof(common.BytesToHash(rootHash), key, proof)
}
