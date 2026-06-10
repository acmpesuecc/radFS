package art

// TODO: Leaf node structure for storing values

type leaf struct {
	key    []byte
	values interface{}
}

func newleaf(value interface{}, key []byte) *Node {
	return &Node{
		leaf: &leaf{key: key, values: value},
	}

}
func isleaf(n *Node) bool {
	return n.leaf != nil
}
