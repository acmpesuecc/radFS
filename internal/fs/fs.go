package fs

import (
	"sync"
	"sync/atomic"

	"bazil.org/fuse/fs"
)

type FS struct{}

var inodeCounter uint64 = 2

func nextInode() uint64 {
	return atomic.AddUint64(&inodeCounter, 1)
}

func (FS) Root() (fs.Node, error) {
	root := &Dir{
		inode: 1,
		Nodes: map[string]fs.Node{
			"hello.txt": &File{
				inode: nextInode(),
				data:  []byte("Hello from radFS!\n"),
				mode:  0o666,
			},
		},
	}
	return root, nil
}

type File struct {
	mu    sync.Mutex
	inode uint64
	data  []byte
	mode  uint32
}

type Dir struct {
	mu    sync.Mutex
	inode uint64
	Nodes map[string]fs.Node
}
