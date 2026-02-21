package fs

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1                 //inode 1 cuz root
	a.Mode = os.ModeDir | 0o755 //octal perms for rwx r-x r-x
	return nil
}
