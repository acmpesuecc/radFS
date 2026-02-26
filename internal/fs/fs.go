package fs

import (
	"bazil.org/fuse/fs"
)

type FS struct{}

var debug = true

func (FS) Root() (fs.Node, error) {
	return &Dir{
		Nodes: map[string]fs.Node{
			"hello.txt": &File{
				Name: "hello.txt",
				Data: []byte("Hello from radFS!\n"),
			}, // Hardcoded default file
		},
	}, nil
}
