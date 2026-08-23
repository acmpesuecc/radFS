package art

import (
	"bytes"
)

// hand over hand locking search function -> lock the current node, find the next node, lock it, then unlock the current node
func search(n *Node, key []byte, depth int) *Node {
	cur := n
	cur.mu.RLock()

	for cur != nil {
		if isleaf(cur) {
			if bytes.Equal(cur.leaf.key, key) {
				cur.mu.RUnlock()
				return cur
			}
			cur.mu.RUnlock()
			return nil
		}

		if cur.innerNode.meta.prefixlen > 0 {
			p := checkprefix(cur, key, depth)

			if p != cur.innerNode.meta.prefixlen {
				cur.mu.RUnlock()

				return nil
			}
			depth += cur.innerNode.meta.prefixlen
		}

		//	KEY EXHAUSTION CHECK (The Fix)
		//
		// If we've consumed the entire key, the value must be in this inner node's leaf
		if depth == len(key) {
			if cur.innerNode.leaf != nil {
				cur.mu.RUnlock()
				return cur.innerNode.leaf
			}
			cur.mu.RUnlock()
			return nil
		}

		k := key[depth]
		next, _ := findchild(k, cur)
		if next == nil {
			cur.mu.RUnlock()

			return nil
		}
		next.mu.RLock()
		cur.mu.RUnlock()

		cur = next
		depth++

	}
	return nil

}
