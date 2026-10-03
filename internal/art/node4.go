package art

func newNode4() *Node {
	in := &innerNode{
		nodeType:     Node4,
		keys:         make([]byte, Node4max),
		children:     make([]*Node, Node4max),
		num_children: 0,
		meta:         meta{},
	}
	return &Node{innerNode: in}

}
