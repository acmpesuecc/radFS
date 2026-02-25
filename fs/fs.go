package fs

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FS struct{ Debug bool }

func (f FS) Root() (fs.Node, error) {
	return &Dir{
		Debug: f.Debug,
		Nodes: map[string]fs.Node{
			"hello.txt": &File{},
		},
	}, nil
}

type Dir struct {
	Debug bool
	Nodes map[string]fs.Node
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {

	a.Inode = 1                 //inode 1 cuz root
	a.Mode = os.ModeDir | 0o755 //octal perms for rwx r-x r-x

	return nil
}
