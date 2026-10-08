package art

import (
	"bytes"
)

func deletekey(n *Node, key []byte, depth int) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	if isleaf(n) {
		if bytes.Equal(n.leaf.key, key) {
			return nil, true
		}
		return n, false
	}
	var parent *Node

	parentbyte := byte(0)

	cur := n
	cur.mu.Lock()
	for {
		p := checkprefix(cur, key, depth)
		if p != cur.innerNode.meta.prefixlen {
			if parent != nil {
				parent.mu.Unlock()
			}
			cur.mu.Unlock()
			return n, false

		}
		depth += cur.innerNode.meta.prefixlen

		// 3. KEY EXHAUSTION (The Fix)
		// If the key ends here, the value is in the inner node's leaf field
		if depth == len(key) {
			if cur.innerNode.leaf == nil {
				if parent != nil {
					parent.mu.Unlock()
				}
				cur.mu.Unlock()
				return n, false
			}

			cur.innerNode.leaf = nil // Remove the value
			if shouldShrink(cur) {   //shirnk the node if needed
				newNode := shrink(cur)
				if parent != nil {
					if newNode == nil { // If the inner node is empty after removing the leaf, remove it from the parent
						parent = removechild(parent, parentbyte)
					} else {
						parent = addchild(parent, parentbyte, newNode)
					}
					parent.mu.Unlock()
					cur.mu.Unlock()
					return n, true
				}
				cur.mu.Unlock()
				return newNode, true

			}

			if parent != nil {
				parent.mu.Unlock()
			}
			cur.mu.Unlock()
			return n, true
		}

		k := key[depth]
		child, _ := findchild(k, cur)
		if child == nil {
			if parent != nil {
				parent.mu.Unlock()
			}
			cur.mu.Unlock()
			return n, false
		}
		child.mu.Lock()

		if isleaf(child) {
			if !bytes.Equal(child.leaf.key, key) {
				child.mu.Unlock()
				if parent != nil {
					parent.mu.Unlock()
				}
				cur.mu.Unlock()
				return n, false
			}

			newcur := removechild(cur, k) // Remove the leaf from the inner node
			child.mu.Unlock()

			if parent != nil {
				if newcur == nil { // If the inner node is empty after removing the leaf, remove it from the parent
					parent = removechild(parent, parentbyte)
				} else {
					parent = addchild(parent, parentbyte, newcur)
				}

				parent.mu.Unlock()
				cur.mu.Unlock()
				return n, true
			}
			cur.mu.Unlock()
			return newcur, true

		}
		// Move down the tree when innernode
		if parent != nil {
			parent.mu.Unlock()
		}
		parent = cur
		parentbyte = k
		cur = child
		depth++

	}
}
