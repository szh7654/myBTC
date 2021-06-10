package BLC

import (
	"crypto/sha256"
)

type Node struct {
	left  *Node  //左子节点
	right *Node  //右子节点
	data  []byte //hash
}

type MerkleTree struct {
	root *Node
}

func NewMerkleNode(left *Node, right *Node, data []byte) *Node {
	node := &Node{}
	var hash [32]byte

	// leaf node
	if left == nil && right == nil {
		hash = sha256.Sum256(data)
	} else {
		// non-leaf node
		hash = sha256.Sum256(append(left.data, right.data...))
	}

	node.left = left
	node.right = right
	node.data = hash[:]
	return node
}

func NewMerkleTree(data [][]byte) *MerkleTree {

	var nodes []*Node

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, data := range data {
		node := NewMerkleNode(nil, nil, data)
		nodes = append(nodes, node)
	}
	// Generate the upper node recursively
	for {
		var newNodes []*Node
		for i := 0; i < len(nodes); i += 2 {
			node := NewMerkleNode(nodes[i], nodes[i+1], nil)
			newNodes = append(newNodes, node)
		}
		nodes = newNodes
		if len(newNodes) == 1 {
			break
		}
	}

	merkleTree := &MerkleTree{nodes[0]}

	return merkleTree
}
