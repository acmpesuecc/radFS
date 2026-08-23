package art

import "sync"

// TODO: Interfaces and shared node header (meta)

type NodeType int

const (
	Node4 NodeType = iota
	Node16
	Node48
	Node256
)
const (
	Node4max   = 4
	Node16Max  = 16
	Node48Max  = 48
	Node256Max = 256

	maxprefixlen = 8
)

type Node struct {
	innerNode *innerNode
	leaf      *leaf
	mu        sync.RWMutex
}

type innerNode struct {
	nodeType NodeType
	keys     []byte
	children []*Node
	leaf     *Node

	num_children int

	meta meta
}

type meta struct {
	prefix    []byte
	prefixlen int
}
