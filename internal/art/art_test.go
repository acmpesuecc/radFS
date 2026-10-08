package art

import (
	"fmt"
	"sync"
	"testing"
)

// go test -race -count=1 -v .  (run with race detector and no caching)
func TestInsertAndSearch(t *testing.T) {
	var tree Tree

	tree.Insert([]byte("hello"), "world")

	v, ok := tree.Search([]byte("hello"))

	if !ok {
		t.Fatal("key not found")
	}
	if v != "world" {
		t.Fatalf("got %v wanted world", v)
	}

}

func TestSearchMissing(t *testing.T) {
	var tree Tree

	_, ok := tree.Search([]byte("ghost"))

	if ok {
		t.Fatal("expected key not found")
	}
}

func TestDelete(t *testing.T) {
	var tree Tree

	tree.Insert([]byte("hello"), "world")

	if !tree.Delete([]byte("hello")) {
		t.Fatal("delete failed")
	}

	_, ok := tree.Search([]byte("hello"))

	if ok {
		t.Fatal("key still exists")
	}
}

func TestPrefixOverlap(t *testing.T) {
	var tree Tree

	keys := []string{
		"a",
		"ab",
		"abc",
		"abcd",
	}

	for _, k := range keys {
		tree.Insert([]byte(k), k)
	}

	for _, k := range keys {
		_, ok := tree.Search([]byte(k))

		if !ok {
			t.Fatalf("missing key %s", k)
		}
	}
}

func TestDeletePrefix(t *testing.T) {
	var tree Tree

	tree.Insert([]byte("a"), "a")
	tree.Insert([]byte("ab"), "ab")
	tree.Insert([]byte("abc"), "abc")

	tree.Delete([]byte("ab"))

	if _, ok := tree.Search([]byte("ab")); ok {
		t.Fatal("ab still exists")
	}

	if _, ok := tree.Search([]byte("a")); !ok {
		t.Fatal("a missing")
	}

	if _, ok := tree.Search([]byte("abc")); !ok {
		t.Fatal("abc missing")
	}
}

func TestUpdate(t *testing.T) {
	var tree Tree

	tree.Insert([]byte("hello"), "goob")

	tree.Insert([]byte("hello"), "new")

	v, _ := tree.Search([]byte("hello"))
	fmt.Printf("%v", v)

	if v != "new" {
		t.Fatalf("got %v want new", v)
	}
}

func TestForEach(t *testing.T) {
	var tree Tree

	tree.Insert([]byte("cat"), "cat")
	tree.Insert([]byte("apple"), "apple")
	tree.Insert([]byte("banana"), "banana")

	var result []string

	tree.ForEach(func(k []byte, v interface{}) {
		result = append(result, string(k))
	})

	expected := []string{
		"apple",
		"banana",
		"cat",
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Fatalf("got %v want %v", result, expected)
		}
	}
}

func TestNodeGrowth(t *testing.T) {
	var tree Tree

	for i := 0; i < 5; i++ {
		tree.Insert([]byte{byte(i)}, "x")
	}

	if GetNodeTypeName(tree.Root()) != "Node16" {
		t.Fatal("expected Node16")
	}
	for i := 0; i < 17; i++ {
		tree.Insert([]byte{byte(i)}, "x")
	}

	if GetNodeTypeName(tree.Root()) != "Node48" {
		t.Fatal("expected Node48")
	}
	for i := 0; i < 49; i++ {
		tree.Insert([]byte{byte(i)}, "x")
	}

	if GetNodeTypeName(tree.Root()) != "Node256" {
		t.Fatal("expected Node256")
	}

}

func TestNodeShrink(t *testing.T) {
	var tree Tree

	for i := 0; i < 49; i++ {
		tree.Insert([]byte{byte(i)}, "x")
	}
	tree.Delete([]byte{byte(48)})
	if GetNodeTypeName(tree.Root()) != "Node48" {
		t.Fatal("expected Node48")
	}
	for i := 47; i > 15; i-- {
		tree.Delete([]byte{byte(i)})

	}
	if GetNodeTypeName(tree.Root()) != "Node16" {
		t.Fatal("expected Node16")
	}
	for i := 15; i > 3; i-- {
		tree.Delete([]byte{byte(i)})

	}
	if GetNodeTypeName(tree.Root()) != "Node4" {
		t.Fatal("expected Node4")
	}

}

func TestConcurrentInsertSearch(t *testing.T) {
	var tree Tree
	var wg sync.WaitGroup

	const numGoroutines = 50
	const keysPerGoroutine = 100

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < keysPerGoroutine; i++ {
				key := fmt.Appendf(nil, "g%d-k%d", g, i)
				tree.Insert(key, g*1000+i)
			}
		}(g)
	}
	wg.Wait()

	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < keysPerGoroutine; i++ {
			key := fmt.Appendf(nil, "g%d-k%d", g, i)
			v, ok := tree.Search(key)
			if !ok {
				t.Fatalf("missing key %s", key)
			}
			if v != g*1000+i {
				t.Fatalf("key %s: got %v want %v", key, v, g*1000+i)
			}
		}
	}
}

func TestConcurrentInsertDeleteSearch(t *testing.T) {
	var tree Tree
	var wg sync.WaitGroup

	const numKeys = 500
	keys := make([][]byte, numKeys)
	for i := range keys {
		keys[i] = fmt.Appendf(nil, "key-%d", i)
		tree.Insert(keys[i], i) // seed sequentially first
	}

	// concurrent deleters (odd-indexed keys) + updates (even-indexed)
	for i := 0; i < numKeys; i++ {
		if i%2 == 1 {
			wg.Add(1)
			go func(k []byte) {
				defer wg.Done()
				tree.Delete(k)
			}(keys[i])
		} else {
			wg.Add(1)
			go func(k []byte, v int) {
				defer wg.Done()
				tree.Insert(k, v+10000) // concurrent update
			}(keys[i], i)
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numKeys; i++ {
			tree.Search(keys[i]) // concurrent search
		}
	}()
	wg.Wait()

	for i := 0; i < numKeys; i++ {
		v, ok := tree.Search(keys[i])
		if i%2 == 1 {
			if ok {
				t.Fatalf("key %s should be deleted, got %v", keys[i], v)
			}
		} else {
			if !ok || v != i+10000 {
				t.Fatalf("key %s: got %v, %v; want %v, true", keys[i], v, ok, i+10000)
			}
		}
	}
}

func TestPrefixSplitAfterStoredPrefix(t *testing.T) {
	var tree Tree

	tree.Insert([]byte("aaaaaaaaX"), 1)
	tree.Insert([]byte("aaaaaaaaY"), 2)
	tree.Insert([]byte("aaaaaaaZ"), 3)

	for _, key := range []string{
		"aaaaaaaaX",
		"aaaaaaaaY",
		"aaaaaaaZ",
	} {
		if _, ok := tree.Search([]byte(key)); !ok {
			t.Fatalf("missing key %q", key)
		}
	}
}
