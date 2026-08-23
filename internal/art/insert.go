package art

func insert(n *Node, value interface{}, key []byte, depth int) *Node {

	if n == nil {

		return newleaf(value, key)
	}
	cur := n
	cur.mu.Lock()

	for {

		if isleaf(cur) {
			new_node := newNode4()
			oldkey := cur.leaf.key
			i := depth

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
				new_node.innerNode.meta.prefix = deepcopy(cur.innerNode.meta.prefix[:p])

			} else {
				new_node.innerNode.meta.prefix = deepcopy(cur.innerNode.meta.prefix[:maxprefixlen])
			}

			oldprefixlen := cur.innerNode.meta.prefixlen
			cur.innerNode.meta.prefixlen = oldprefixlen - (p + 1)
			if oldprefixlen < maxprefixlen {
				cur.innerNode.meta.prefix = deepcopy(cur.innerNode.meta.prefix[p+1 : oldprefixlen])

			} else {
				leaf := fetchleaf(cur)
				cur.innerNode.meta.prefix = deepcopy(leaf.leaf.key[depth+p+1 : depth+p+1+maxprefixlen])
			}
			cur.mu.Unlock()

			return new_node
		}

		depth += cur.innerNode.meta.prefixlen
		if depth == len(key) {
			cur.innerNode.leaf = newleaf(value, key)
			cur.mu.Unlock()
			return cur
		}
		next, _ := findchild(key[depth], cur)
		if next == nil {
			oldcur := cur

			newcur := addchild(cur, key[depth], newleaf(value, key))
			oldcur.mu.Unlock()
			return newcur

		}
		next.mu.Lock()
		cur.mu.Unlock()
		cur = next
		depth++

	}

}
