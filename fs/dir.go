package fs

import (
	"context"
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var debug = true


func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if debug {
		log.Printf("Lookup called for: %s", name)
	}
	return nil, fuse.ENOENT
}


func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if debug {
		log.Println("ReadDirAll() called")
	}
	return []fuse.Dirent{}, nil
}
