package art

// TODO: Node16 implementation
func newNode16() *Node {
	in := &innerNode{
		nodeType:     Node16,
		keys:         make([]byte, Node16Max),
		children:     make([]*Node, Node16Max),
		num_children: 0,
		meta: meta{
			prefix: make([]byte, maxprefixlen),
		},
	}
	return &Node{innerNode: in}
}
