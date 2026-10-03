package art

import "math/bits"

const node48FullMask = (uint64(1) << 48) - 1

// TODO: Helper functions (e.g., prefix matching)
func addchild(n *Node, k byte, child *Node) *Node {
	in := n.innerNode

	child1, pos1 := findchild(k, n) // to prevent duplicate insertions

	if child1 != nil {
		in.children[pos1] = child
		return n
	}
	switch n.innerNode.nodeType {
	case Node16, Node4:
		if n.innerNode.num_children == len(in.keys) {
			n = grow(n)

			return addchild(n, k, child)

		}
		var i int
		for i = in.num_children - 1; i >= 0 && k < in.keys[i]; i-- { //shifts until keybyte place is found
			in.keys[i+1] = in.keys[i]
			in.children[i+1] = in.children[i]

		}

		in.keys[i+1] = k
		in.children[i+1] = child
		in.num_children++

		return n
	case Node48:
		if n.innerNode.num_children == len(in.children) {
			n = grow(n)

			return addchild(n, k, child)

		}
		if in.keys[k] != 0 { // if key exist then update
			key := int(n.innerNode.keys[k]) - 1 // the zero slot is used to check if its an empty key so we start filling the index values in key from 1
			n.innerNode.children[key] = child
			return n

		}
		i := bits.TrailingZeros64(in.freeMask) // first free slot, O(1)
		in.children[i] = child
		in.keys[k] = byte(i + 1)
		in.freeMask &^= 1 << uint(i) // mark slot occupied
		in.num_children++

	case Node256:
		if in.children[k] != nil { //update key
			in.children[k] = child
			return n

		}
		in.children[k] = child //inserting new key
		in.num_children++

	}

	return n

}
func checkprefix(n *Node, key []byte, depth int) int {
	in := n.innerNode
	var i int
	maxcmp := min(maxprefixlen, in.meta.prefixlen)

	for i = 0; i < maxcmp && depth+i < len(key); i++ { //checks prefix until mismatch
		if in.meta.prefix[i] != key[depth+i] {
			return i // case when you find mismatch and the mismatch is less than maxprefixlen

		}

	}
	if in.meta.prefixlen > maxprefixlen {
		leaf := fetchleaf(n)
		leafkey := leaf.leaf.key
		for ; i < in.meta.prefixlen && depth+i < len(leafkey) && depth+i < len(key); i++ {
			if key[depth+i] != leaf.leaf.key[depth+i] {
				return i // case when you find mismatch and the mismatch is more than maxprefixlen

			}

		}

	}

	return i // case when you find mismatch and the mismatch is equal maxprefixlen

}

func findchild(k byte, n *Node) (*Node, int) {
	in := n.innerNode
	switch in.nodeType {
	case Node4, Node16:
		for i := 0; i < in.num_children; i++ {
			if in.keys[i] == k {
				return in.children[i], i //finds the node and the position
			}

		}
	case Node48:
		idx := in.keys[k]
		if idx > 0 {
			realindex := int(idx - 1)
			return in.children[realindex], realindex

		}
	case Node256:
		if in.children[k] != nil {
			return in.children[k], int(k)
		}

	}
	return nil, -1

}

func removechild(n *Node, k byte) *Node {
	in := n.innerNode
	_, pos := findchild(k, n)

    // If child doesn't exist, nothing to remove: return the original node unchanged
	if pos == -1  {
		return n
	}

	switch in.nodeType {
	case Node4, Node16:

		for i := pos; i < in.num_children-1; i++ {
			in.keys[i] = in.keys[i+1]
			in.children[i] = in.children[i+1]
		}
		in.keys[in.num_children-1] = 0
		in.children[in.num_children-1] = nil
		in.num_children--

	case Node48:

		idx := in.keys[k]
		if idx > 0 {
			in.keys[k] = 0
			in.children[idx-1] = nil
			in.freeMask |= 1 << uint(idx-1) // mark slot free
			in.num_children--
		}

	case Node256:
		if in.children[k] != nil {
			in.children[k] = nil
			in.num_children--

		}

	}

	if shouldShrink(n) {
		return shrink(n)
	}

	return n
}

func shouldShrink(n *Node) bool {
	in := n.innerNode
	switch in.nodeType {
	case Node256:
		return in.num_children <= 48
	case Node48:
		return in.num_children <= 16
	case Node16:
		return in.num_children <= 4
	case Node4:
		return in.num_children <= 1
	}
	return false
}

func grow(n *Node) *Node {
	switch n.innerNode.nodeType {
	case Node4:
		n16 := newNode16()
		copymeta(n, n16)
		index := 0
		for i := 0; i < 4; i++ {
			if n.innerNode.children[i] != nil {
				n16.innerNode.keys[index] = n.innerNode.keys[i]
				n16.innerNode.children[index] = n.innerNode.children[i]
				index++

			}

		}
		n16.innerNode.num_children = index
		return n16
	case Node16:
		n48 := newNode48()
		copymeta(n, n48)
		index := 0
		for i := 0; i < n.innerNode.num_children; i++ {
			idx := n.innerNode.keys[i]
			child := n.innerNode.children[i]

			if child != nil {

				n48.innerNode.keys[idx] = byte(index + 1)
				n48.innerNode.children[index] = child
				index++

			}

		}
		n48.innerNode.num_children = index
		n48.innerNode.freeMask = node48FullMask &^ ((uint64(1) << uint(index)) - 1) //slots from 0 to index-1 are occupied (mark them as occupied in freeMask)
		return n48

	case Node48:
		n256 := newNode256()
		copymeta(n, n256)
		count := 0
		for i := 0; i < 256; i++ {
			idx := n.innerNode.keys[i]

			if n.innerNode.keys[i] != 0 {
				child := n.innerNode.children[int(idx-1)]
				n256.innerNode.children[i] = child
				count++
			}

		}
		n256.innerNode.num_children = count

		return n256

	}
	return nil

}

func shrink(n *Node) *Node {
	in := n.innerNode
	switch in.nodeType {
	case Node4:
		if in.num_children == 0 {
			if in.leaf != nil {
				return in.leaf
			}

			return nil
		}
		// when old node has only one child and no leaf, we can merge the prefix of the old node with the edge byte and the prefix of the child node into the child node. This reduces the height of the tree and improves search performance.
		// eg ant, ante then anteater , delete anteater, then the prefix of ante and the edge byte 'e' and the prefix of eater can be merged into eater node. This reduces the height of the tree and improves search performance.
		//Fold n's prefix + edgeByte + child's existing prefix into child
		if in.num_children == 1 && in.leaf == nil {
			child := in.children[0]
			edgebyte := in.keys[0]
			if isleaf(child) {
				return child
			}
			oldchildprefixlen := child.innerNode.meta.prefixlen
			new_prefix := make([]byte, 0, maxprefixlen)
			new_prefix = append(new_prefix, in.meta.prefix[:min(in.meta.prefixlen, maxprefixlen)]...)
			if len(new_prefix) < maxprefixlen {
				new_prefix = append(new_prefix, edgebyte)
			}
			if len(new_prefix) < maxprefixlen {
				remaining := maxprefixlen - len(new_prefix)
				new_prefix = append(new_prefix, child.innerNode.meta.prefix[:min(remaining, oldchildprefixlen)]...)
			}

			child.innerNode.meta.prefix = new_prefix
			child.innerNode.meta.prefixlen = in.meta.prefixlen + 1 + oldchildprefixlen

			return child
		}

		return n

	case Node16:
		n4 := newNode4()
		copymeta(n, n4)
		for i := 0; i < in.num_children; i++ {
			n4.innerNode.keys[i] = in.keys[i]
			n4.innerNode.children[i] = in.children[i]
		}
		n4.innerNode.num_children = in.num_children
		return n4

	case Node48:
		n16 := newNode16()
		copymeta(n, n16)
		count := 0
		for i := 0; i < 256; i++ {
			idx := in.keys[i]
			if idx > 0 {
				n16.innerNode.keys[count] = byte(i)
				n16.innerNode.children[count] = in.children[idx-1]
				count++
			}
		}
		n16.innerNode.num_children = count
		return n16

	case Node256:
		n48 := newNode48()
		copymeta(n, n48)
		count := 0
		for i := 0; i < 256; i++ {
			child := in.children[i]
			if child != nil {
				n48.innerNode.children[count] = child
				n48.innerNode.keys[byte(i)] = byte(count + 1)
				count++
			}
		}
		n48.innerNode.num_children = count
		n48.innerNode.freeMask = node48FullMask &^ ((uint64(1) << uint(count)) - 1) //slots from 0 to count-1 are occupied (mark them as occupied in freeMask)
		return n48
	}
	return n
}

func copymeta(n *Node, new_node *Node) {
	new_node.innerNode.meta.prefixlen = n.innerNode.meta.prefixlen
	copy(new_node.innerNode.meta.prefix, n.innerNode.meta.prefix) // both are maxprefixlen long
	new_node.innerNode.leaf = n.innerNode.leaf
}

func fetchleaf(n *Node) *Node {
	if isleaf(n) {
		return n
	}
	if n.innerNode.leaf != nil {
		return n.innerNode.leaf
	}

	in := n.innerNode
	switch in.nodeType {
	case Node4, Node16:
		for i := 0; i < in.num_children; i++ {
			if in.children[i] != nil {
				return fetchleaf(in.children[i])
			}
		}
	case Node48:
		for _, child := range in.children {
			if child != nil {
				return fetchleaf(child)
			}
		}
	case Node256:
		for i := 0; i < len(in.children); i++ {
			if in.children[i] != nil {
				return fetchleaf(in.children[i])
			}
		}
	}
	return nil

}

