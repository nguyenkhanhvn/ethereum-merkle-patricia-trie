package nodes

import (
	"github.com/ethereum/go-ethereum/rlp"
)

type Node interface {
	Hash() ([]byte, error) // common.Hash
	Raw() ([]interface{}, error)
}

func Hash(node Node) ([]byte, error) {
	if IsEmptyNode(node) {
		return EmptyNodeHash, nil
	}
	return node.Hash()
}

func Serialize(node Node) ([]byte, error) {
	var raw interface{}
	var err error

	if IsEmptyNode(node) {
		raw = EmptyNodeRaw
	} else {
		raw, err = node.Raw()
		if err != nil {
			return nil, err
		}
	}

	rlp, err := rlp.EncodeToBytes(raw)
	if err != nil {
		return nil, err
	}

	return rlp, nil
}
