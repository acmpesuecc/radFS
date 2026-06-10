package art

// TODO: Node256 implementation (direct map)

func newNode256() *Node {
	in := &innerNode{
		nodeType:     Node256,
		children:     make([]*Node, Node256Max),
		num_children: 0,

		meta: meta{
			prefix: make([]byte, maxprefixlen),
		},
	}
	return &Node{innerNode: in}
}
