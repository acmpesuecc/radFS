package art

func insert(n *Node, value string, key []byte, depth int) *Node {

	if n == nil {
		return newleaf(value, key)
	}
	if isleaf(n) {
		new_node := newNode4()
		oldkey := n.leaf.key
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
			new_node.innerNode.leaf = n

		} else {
			new_node = addchild(new_node, oldkey[depth], n)

		}

		return new_node

	}
	p := checkprefix(n, key, depth)

	if p != n.innerNode.meta.prefixlen {

		new_node := newNode4()
		if p+depth == len(key) {
			new_node.innerNode.leaf = newleaf(value, key)

		} else {
			new_node = addchild(new_node, key[depth+p], newleaf(value, key))

		}
		leaf := fetchleaf(n) // either its an actual leaf or innernode leaf
		oldkey := leaf.leaf.key

		var oldkeybyte byte
		if p < maxprefixlen {
			oldkeybyte = n.innerNode.meta.prefix[p]
		} else {
			oldkeybyte = oldkey[depth+p]
		}

		new_node = addchild(new_node, oldkeybyte, n)

		new_node.innerNode.meta.prefixlen = p
		if p < maxprefixlen {
			new_node.innerNode.meta.prefix = deepcopy(n.innerNode.meta.prefix[:p])

		} else {
			new_node.innerNode.meta.prefix = deepcopy(n.innerNode.meta.prefix[:maxprefixlen])
		}

		oldprefixlen := n.innerNode.meta.prefixlen
		n.innerNode.meta.prefixlen = oldprefixlen - (p + 1)
		if oldprefixlen < maxprefixlen {
			n.innerNode.meta.prefix = deepcopy(n.innerNode.meta.prefix[p+1 : oldprefixlen])

		} else {
			leaf := fetchleaf(n)
			leafKey := leaf.leaf.key
			start := depth + p + 1
			if start >= len(leafKey) {
				n.innerNode.meta.prefix = []byte{}
			} else {
				end := start + maxprefixlen
				if end > len(leafKey) {
					end = len(leafKey)
				}
				n.innerNode.meta.prefix = deepcopy(leafKey[start:end])
			}
		}

		return new_node
	}

	depth += n.innerNode.meta.prefixlen
	if depth == len(key) {
		n.innerNode.leaf = newleaf(value, key)
		return n
	}
	next, pos := findchild(key[depth], n)
	if next != nil {
		n.innerNode.children[pos] = insert(next, value, key, depth+1)
		return n

	} else {
		n = addchild(n, key[depth], newleaf(value, key))
		return n

	}

}
