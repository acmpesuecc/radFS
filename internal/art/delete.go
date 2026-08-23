package art

func deletekey(n *Node, key []byte, depth int) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	if isleaf(n) {
		if string(n.leaf.key) == string(key) {
			return nil, true
		}
		return n, false
	}

	p := checkprefix(n, key, depth)
	if p != n.innerNode.meta.prefixlen {
		return n, false
	}
	depth += n.innerNode.meta.prefixlen

	// 3. KEY EXHAUSTION (The Fix)
	// If the key ends here, the value is in the inner node's leaf field
	if depth == len(key) {
		if n.innerNode.leaf != nil {
			n.innerNode.leaf = nil // Remove the value
			return n, true
		}
		return n, false
	}

	k := key[depth]
	child, pos := findchild(k, n)
	if child == nil {
		return n, false
	}

	newChild, deleted := deletekey(child, key, depth+1)
	if !deleted {
		return n, false
	}

	if newChild == nil {
		n = removechild(n, k)
	} else {
		n.innerNode.children[pos] = newChild
	}

	return n, true
}
