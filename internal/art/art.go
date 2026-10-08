package art

import "sync"

// TODO: Public API (Tree struct, Insert, Search, Delete)

type Tree struct {
	root *Node
	mu   sync.RWMutex
}

func (t *Tree) Root() *Node {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

func (t *Tree) Insert(key []byte, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = insert(t.root, value, key, 0)
}

func (t *Tree) Search(key []byte) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	leaf := search(t.root, key, 0) // start from root and depth 0
	if leaf != nil && isleaf(leaf) {
		return leaf.leaf.values, true //Node->innerleaf->values
	}
	return nil, false
}
func (t *Tree) Delete(key []byte) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
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
	t.mu.RLock()
	defer t.mu.RUnlock()
	traverse(t.root, 0, fn)
}

func New() *Tree { return &Tree{} }

func (t *Tree) Empty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root == nil
}
