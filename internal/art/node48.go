package art

// TODO: Node48 implementation (indirection layer)
func newNode48() *Node {
	in := &innerNode{
		nodeType:     Node48,
		keys:         make([]byte, Node256Max),
		children:     make([]*Node, Node48Max),
		num_children: 0,
		meta:         meta{},
		freeMask:     node48FullMask,
	}

	return &Node{innerNode: in}
}
