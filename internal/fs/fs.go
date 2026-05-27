package fs

import (
	"sync"
	"sync/atomic"

	"bazil.org/fuse/fs"
	"github.com/acmpesuecc/radFS/internal/art"
)

type FS struct {
	Debug bool
}

var inodeCounter uint64 = 2

func nextInode() uint64 {
	return atomic.AddUint64(&inodeCounter, 1)
}

func (f *FS) Root() (fs.Node, error) {
	root := &Dir{
		inode: 1,
		tree:  art.New(),

		fs: f,
	}
	hello := &File{
		inode: nextInode(),
		data:  []byte("Hello from radFS!\n"),
		mode:  0o666,
	}

	root.tree.Insert([]byte("hello.txt"), hello)

	return root, nil
}

type File struct {
	mu    sync.RWMutex
	inode uint64
	data  []byte
	mode  uint32
}

type Dir struct {
	mu    sync.RWMutex
	inode uint64
	tree  *art.Tree
	fs    *FS
}
