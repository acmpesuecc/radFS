package art

import (
	"fmt"
	"testing"
)

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
