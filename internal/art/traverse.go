package art

func traverse(n *Node, depth int, fn func([]byte, interface{})) {
	if n == nil {
		return
	}

	if isleaf(n) {
		fn(n.leaf.key, n.leaf.values)
		return
	}

	in := n.innerNode

	prefixlen := in.meta.prefixlen

	if in.leaf != nil {
		fn(in.leaf.leaf.key, in.leaf.leaf.values)
	}

	newDepth := depth + prefixlen

	switch in.nodeType {
	case Node4, Node16:
		for i := 0; i < in.num_children; i++ {

			child := in.children[i]

			traverse(child, newDepth+1, fn)
		}

	case Node48:
		for b := 0; b < 256; b++ {
			idx := in.keys[b]

			if idx != 0 {

				child := in.children[idx-1]

				traverse(child, newDepth+1, fn)
			}
		}

	case Node256:
		for b := 0; b < 256; b++ {
			child := in.children[b]
			if child != nil {
				traverse(child, newDepth+1, fn)
			}
		}
	}
}
