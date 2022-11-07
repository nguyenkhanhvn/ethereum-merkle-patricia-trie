package trie

import (
	"fmt"

	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/nodes"
	"github.com/nguyenkhanhvn/ethereum-merkle-patricia-trie/utils"
)

type Trie struct {
	root nodes.Node
}

func NewTrie() *Trie {
	return &Trie{}
}

func (t *Trie) Hash() ([]byte, error) {
	if nodes.IsEmptyNode(t.root) {
		return nodes.EmptyNodeHash, nil
	}
	return t.root.Hash()
}

func (t *Trie) Get(key []byte) ([]byte, bool) {
	node := t.root
	nibbles := utils.FromBytes(key)
	for {
		if nodes.IsEmptyNode(node) {
			return nil, false
		}

		if leaf, ok := node.(*nodes.LeafNode); ok {
			matched := utils.PrefixMatchedLen(leaf.Path, nibbles)
			if matched != len(leaf.Path) || matched != len(nibbles) {
				return nil, false
			}
			return leaf.Value, true
		}

		if branch, ok := node.(*nodes.BranchNode); ok {
			if len(nibbles) == 0 {
				return branch.Value, branch.HasValue()
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
				return nil, false
			}

			nibbles = nibbles[matched:]
			node = ext.Next
			continue
		}

		return nil, false
	}
}

// Put adds a key value pair to the trie
// In general, the rule is:
// - When stopped at an EmptyNode, replace it with a new LeafNode with the remaining path.
// - When stopped at a LeafNode, convert it to an ExtensionNode and add a new branch and a new LeafNode.
// - When stopped at an ExtensionNode, convert it to another ExtensionNode with shorter path and create a new BranchNode points to the ExtensionNode.
func (t *Trie) Put(key []byte, value []byte) error {
	// need to use pointer, so that I can update root in place without
	// keeping trace of the parent node
	node := &t.root
	nibbles := utils.FromBytes(key)
	for {
		if nodes.IsEmptyNode(*node) {
			leaf := nodes.NewLeafNodeFromNibbles(nibbles, value)
			*node = leaf
			return nil
		}

		if leaf, ok := (*node).(*nodes.LeafNode); ok {
			matched := utils.PrefixMatchedLen(leaf.Path, nibbles)

			// if all matched, update value even if the value are equal
			if matched == len(nibbles) && matched == len(leaf.Path) {
				newLeaf := nodes.NewLeafNodeFromNibbles(leaf.Path, value)
				*node = newLeaf
				return nil
			}

			branch := nodes.NewBranchNode()
			// if matched some nibbles, check if matches either all remaining nibbles
			// or all leaf nibbles
			if matched == len(leaf.Path) {
				branch.SetValue(leaf.Value)
			}

			if matched == len(nibbles) {
				branch.SetValue(value)
			}

			// if there is matched nibbles, an extension node will be created
			if matched > 0 {
				// create an extension node for the shared nibbles
				ext := nodes.NewExtensionNode(leaf.Path[:matched], branch)
				*node = ext
			} else {
				// when there no matched nibble, there is no need to keep the extension node
				*node = branch
			}

			if matched < len(leaf.Path) {
				// have dismatched
				// L 01020304 hello
				// + 010203   world

				// 01020304, 0, 4
				branchNibble, leafNibbles := leaf.Path[matched], leaf.Path[matched+1:]
				newLeaf := nodes.NewLeafNodeFromNibbles(leafNibbles, leaf.Value) // not :matched+1
				branch.SetBranch(branchNibble, newLeaf)
			}

			if matched < len(nibbles) {
				// L 01020304 hello
				// + 010203040 world

				// L 01020304 hello
				// + 010203040506 world
				branchNibble, leafNibbles := nibbles[matched], nibbles[matched+1:]
				newLeaf := nodes.NewLeafNodeFromNibbles(leafNibbles, value)
				branch.SetBranch(branchNibble, newLeaf)
			}

			return nil
		}

		if branch, ok := (*node).(*nodes.BranchNode); ok {
			if len(nibbles) == 0 {
				branch.SetValue(value)
				return nil
			}

			b, remaining := nibbles[0], nibbles[1:]
			nibbles = remaining
			node = &branch.Branches[b]
			continue
		}

		// E 01020304
		// B 0 hello
		// L 506 world
		// + 010203 good
		if ext, ok := (*node).(*nodes.ExtensionNode); ok {
			matched := utils.PrefixMatchedLen(ext.Path, nibbles)
			if matched < len(ext.Path) {
				// E 01020304
				// + 010203 good
				extNibbles, branchNibble, extRemainingnibbles := ext.Path[:matched], ext.Path[matched], ext.Path[matched+1:]
				branch := nodes.NewBranchNode()
				if len(extRemainingnibbles) == 0 {
					// E 0102030
					// + 010203 good
					branch.SetBranch(branchNibble, ext.Next)
				} else {
					// E 01020304
					// + 010203 good
					newExt := nodes.NewExtensionNode(extRemainingnibbles, ext.Next)
					branch.SetBranch(branchNibble, newExt)
				}

				if matched < len(nibbles) {
					nodeBranchNibble, nodeLeafNibbles := nibbles[matched], nibbles[matched+1:]
					remainingLeaf := nodes.NewLeafNodeFromNibbles(nodeLeafNibbles, value)
					branch.SetBranch(nodeBranchNibble, remainingLeaf)
				} else if matched == len(nibbles) {
					branch.SetValue(value)
				} else {
					return fmt.Errorf("too many matched (%v > %v)", matched, len(nibbles))
				}

				// if there is no shared extension nibbles any more, then we don't need the extension node
				// any more
				// E 01020304
				// + 1234 good
				if len(extNibbles) == 0 {
					*node = branch
				} else {
					// otherwise create a new extension node
					*node = nodes.NewExtensionNode(extNibbles, branch)
				}
				return nil
			}

			nibbles = nibbles[matched:]
			node = &ext.Next
			continue
		}

		return fmt.Errorf("unknown type")
	}

}
