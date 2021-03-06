// Copyright (c) 2014-2018 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package avl

// find a specific item
func (tree *Tree) Search(key item) (*Node, int) {
	return search(key, tree.root, 0)
}

func search(key item, tree *Node, index int) (*Node, int) {
	if nil == tree {
		return nil, -1
	}

	switch tree.key.Compare(key) {
	case +1: // tree.key > key
		return search(key, tree.left, index)
	case -1: // tree.key < key
		return search(key, tree.right, index+tree.leftNodes+1)
	default:
		return tree, index + tree.leftNodes
	}
}
