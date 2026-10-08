package art

import "bytes"

func insert(n *Node, value interface{}, key []byte, depth int) *Node {

	if n == nil {
		return newleaf(value, key)
	}
	var parent *Node
	parentbyte := byte(0)

	cur := n
	cur.mu.Lock()

	for {
		if isleaf(cur) {
			if bytes.Equal(cur.leaf.key, key) { // exact key match: update in place
				cur.leaf.values = value
				if parent != nil {
					parent.mu.Unlock()
				}
				cur.mu.Unlock()

				return n
			}

			new_node := newNode4()
			oldkey := cur.leaf.key
			i := depth
			if bytes.Equal(oldkey, key) { // if the key already exists, update the value
				cur.leaf.values = value
				if parent != nil {
					parent.mu.Unlock()
				}
				cur.mu.Unlock()

				return n
			}

			for i < len(oldkey) && i < len(key) && oldkey[i] == key[i] {
				prefix_index := i - depth
				if prefix_index < maxprefixlen {
					new_node.innerNode.meta.prefix[prefix_index] = key[i] // stores only the till max prefix
				}

				i++
			}

			new_node.innerNode.meta.prefixlen = i - depth // stores full prefix len even after maxprefixlen
			depth = i
			if depth == len(key) {
				new_node.innerNode.leaf = newleaf(value, key)
			} else {
				new_node = addchild(new_node, key[depth], newleaf(value, key))
			}

			if depth == len(oldkey) {
				new_node.innerNode.leaf = cur
			} else {
				new_node = addchild(new_node, oldkey[depth], cur)
			}

			if parent != nil {
				parent = addchild(parent, parentbyte, new_node)
				parent.mu.Unlock()
				cur.mu.Unlock()

				return n
			}
			cur.mu.Unlock()

			return new_node
		}

		p := checkprefix(cur, key, depth) //inner node prefix check

		if p != cur.innerNode.meta.prefixlen {
			new_node := newNode4()
			if p+depth == len(key) {
				new_node.innerNode.leaf = newleaf(value, key)
			} else {
				new_node = addchild(new_node, key[depth+p], newleaf(value, key))
			}
			leaf := fetchleaf(cur) // either its an actual leaf or innernode leaf
			oldkey := leaf.leaf.key

			var oldkeybyte byte
			if p < maxprefixlen {
				oldkeybyte = cur.innerNode.meta.prefix[p]
			} else {
				oldkeybyte = oldkey[depth+p]
			}

			new_node = addchild(new_node, oldkeybyte, cur)

			new_node.innerNode.meta.prefixlen = p
			if p < maxprefixlen {
				copy(new_node.innerNode.meta.prefix[:], cur.innerNode.meta.prefix[:p])
			} else {
				copy(new_node.innerNode.meta.prefix[:], cur.innerNode.meta.prefix[:maxprefixlen])
			}

			oldprefixlen := cur.innerNode.meta.prefixlen
			newprefixlen := oldprefixlen - (p + 1)
			var shiftedPrefix [maxprefixlen]byte
			copyLen := min(maxprefixlen, newprefixlen)
			copy(
				shiftedPrefix[:copyLen],
				oldkey[depth+p+1:depth+p+1+copyLen],
			)

			cur.innerNode.meta.prefix = shiftedPrefix
			cur.innerNode.meta.prefixlen = newprefixlen

			if parent != nil {
				parent = addchild(parent, parentbyte, new_node)
				parent.mu.Unlock()
				cur.mu.Unlock()

				return n
			}
			cur.mu.Unlock()
			return new_node
		}

		depth += cur.innerNode.meta.prefixlen
		if depth == len(key) {
			cur.innerNode.leaf = newleaf(value, key)
			if parent != nil {
				parent.mu.Unlock()
			}
			cur.mu.Unlock()
			return n
		}
		next, _ := findchild(key[depth], cur)
		if next == nil {
			newcur := addchild(cur, key[depth], newleaf(value, key))
			if parent != nil {
				addchild(parent, parentbyte, newcur)
				parent.mu.Unlock()
				cur.mu.Unlock()
				return n
			}
			cur.mu.Unlock()
			return newcur
		}
		next.mu.Lock()
		if parent != nil {
			parent.mu.Unlock()
		}

		parent = cur
		cur = next
		parentbyte = key[depth] //stores the parent byte to be used in addchild when we need to replace the child

		depth++
	}
}
