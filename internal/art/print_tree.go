package art

import "fmt"

func PrintTree(n *Node, level int, depth int) {
	if n == nil {
		return
	}

	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	if isleaf(n) {
		fmt.Println(indent + "Leaf: " + string(n.leaf.key))
		return
	}

	in := n.innerNode

	prefixlen := in.meta.prefixlen
	prefix := ""
	if prefixlen <= maxprefixlen {
		prefix = string(in.meta.prefix[:prefixlen])

	} else {
		leaf := fetchleaf(n)
		prefix = string(leaf.leaf.key[depth : depth+prefixlen])
	}

	fmt.Println(indent+"Node(prefix=\""+prefix+"\", prefixLen=", prefixlen, ")")

	if in.leaf != nil {
		fmt.Printf("%s  [Internal Leaf]: %s\n", indent, string(in.leaf.leaf.key))
	}
	// Print children

	newDepth := depth + prefixlen

	switch in.nodeType {
	case Node4, Node16:
		for i := 0; i < in.num_children; i++ {
			key := in.keys[i]
			child := in.children[i]

			fmt.Printf("%s Edge('%c' | %d):\t", indent, key, key)
			PrintTree(child, level+1, newDepth+1)
		}

	case Node48:
		for b := 0; b < 256; b++ {
			idx := in.keys[b]

			if idx != 0 {

				child := in.children[idx-1]

				fmt.Printf("%s Edge('%c' | %d):\t", indent, byte(b), b)
				PrintTree(child, level+1, newDepth+1)
			}
		}

	case Node256:
		for b := 0; b < 256; b++ {
			child := in.children[b]
			if child != nil {
				fmt.Printf("%s Edge('%c' | %d):\t", indent, byte(b), b)
				PrintTree(child, level+1, newDepth+1)
			}
		}
	}
}
