package art

import (
	"bytes"
)

func search(n *Node, key []byte, depth int) *Node {
	if n == nil {
		return nil
	}

	if isleaf(n) {
		if bytes.Equal(n.leaf.key, key) {
			return n
		}
		return nil
	}

	if n.innerNode.meta.prefixlen > 0 {
		p := checkprefix(n, key, depth)

		if p != n.innerNode.meta.prefixlen {
			return nil
		}
		depth += n.innerNode.meta.prefixlen
	}

	//  KEY EXHAUSTION CHECK (The Fix)
	// If we've consumed the entire key, the value must be in this inner node's leaf
	if depth == len(key) {
		if n.innerNode.leaf != nil {
			return n.innerNode.leaf
		}
		return nil
	}

	k := key[depth]
	next, _ := findchild(k, n)
	if next != nil {
		return search(next, key, depth+1)
	}

	return nil
}
