package main

import (
	"fmt"

	"github.com/acmpesuecc/radFS/internal/art"
)

func main() {
	t := art.New()

	fmt.Println("====== PHASE 1: STRESS TESTING GROWTH (4 -> 16 -> 48 -> 256) ======")

	// We use keys with a single byte difference to ensure they all go into the SAME internal node
	for i := 0; i < 256; i++ {
		key := []byte{byte(i)}
		t.Insert(key, fmt.Sprintf("%d", i))

		switch i + 1 {

		case 5:
			fmt.Printf("[Check] Inserted 5 keys.  (Current: %d children)\n", i+1)
			fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))
		case 17:
			fmt.Printf("[Check] Inserted 17 keys.  (Current: %d children)\n", i+1)
			fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))
		case 49:
			fmt.Printf("[Check] Inserted 49 keys. (Current: %d children)\n", i+1)
			fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))
		}
	}

	fmt.Printf("Final Growth State: %d children in root.\n", 256)

	fmt.Println("\n===== PHASE 2: SEARCH VERIFICATION =====")
	// Verify we didn't lose data during the messy pointer copying in grow/shrink
	testKeys := []int{0, 15, 47, 100, 255}
	for _, tk := range testKeys {
		val, ok := t.Search([]byte{byte(tk)})
		fmt.Printf("Searching for key %d: Found=%v, Value=%v\n", tk, ok, val)
	}

	fmt.Println("\n===== PHASE 3: STRESS TESTING SHRINK (256 -> 48 -> 16 -> 4) =====")

	for i := 255; i >= 37; i-- {
		t.Delete([]byte{byte(i)})
	}
	fmt.Println("[Check] Deleted down to 37 keys")
	fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))

	for i := 36; i >= 12; i-- {
		t.Delete([]byte{byte(i)})
	}
	fmt.Println("[Check] Deleted down to 12 keys. ")
	fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))

	for i := 11; i >= 3; i-- {
		t.Delete([]byte{byte(i)})
	}
	fmt.Println("[Check] Deleted down to 3 keys. ")
	fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))

	fmt.Println("\n===== PHASE 4: FULL COLLAPSE (Node4 -> Leaf -> Nil) =====")

	// Delete until only 1 key remains (should collapse Node4 into a Leaf)
	for i := 2; i >= 1; i-- {
		t.Delete([]byte{byte(i)})
	}
	fmt.Println("[Check] Deleted down to 1 key. Tree should be a single Leaf (no Internal Node).")
	art.PrintTree(t.Root(), 0, 0)

	// Delete the very last key
	t.Delete([]byte{byte(0)})
	fmt.Println("[Check] Deleted last key. Root should be nil.")

	if t.Root() == nil {
		fmt.Println("SUCCESS: Tree is fully empty (Root is nil).")
	} else {
		fmt.Println("WARNING: Root is not nil. Check your Node4 -> Leaf collapse logic.")
		art.PrintTree(t.Root(), 0, 0)
	}

	fmt.Println("\n===== FINAL TREE STRUCTURE =====")
	fmt.Printf("NodeType: %s\n", art.GetNodeTypeName(t.Root()))

}
