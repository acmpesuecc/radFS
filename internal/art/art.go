package art

// TODO: Public API (Tree struct, Insert, Search, Delete)

type Tree struct {
	root *Node
}

func (t *Tree) Root() *Node {
	return t.root
}

func (t *Tree) Insert(key []byte, value interface{}) {
	t.root = insert(t.root, value, key, 0)
}

func (t *Tree) Search(key []byte) (interface{}, bool) {
	leaf := search(t.root, key, 0) // start from root and depth 0
	if leaf != nil && isleaf(leaf) {
		return leaf.leaf.values, true //Node->innerleaf->values
	}
	return "", false
}
func (t *Tree) Delete(key []byte) bool {
	if t.root == nil {
		return false
	}

	newRoot, deleted := deletekey(t.root, key, 0)

	if deleted {
		t.root = newRoot
		return true
	}

	return false
}

func GetNodeTypeName(n *Node) string {
	if n == nil {
		return "Nil"
	}
	if isleaf(n) {
		return "Leaf"
	}

	types := []string{"Node4", "Node16", "Node48", "Node256"}
	return types[n.innerNode.nodeType]
}
func (t *Tree) ForEach(fn func([]byte, interface{})) {
	traverse(t.root, 0, fn)
}
func New() *Tree {
	return &Tree{}
}
func (t *Tree) Empty() bool {
	return t.root == nil
}
