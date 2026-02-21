package fs

import (
	"context"
	"log"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var debug = true

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if debug {
		log.Printf("Lookup called for: %s", name)
	}

	node, exists := d.Nodes[name]
	if !exists {
		return nil, syscall.ENOENT
	}

	return node, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if debug {
		log.Println("ReadDirAll() called")
	}

	var entries []fuse.Dirent

	// Get all nodes (files/dirs) inside the mountpoint and append to dirents list
	// Not really useful rn because only 1 hardcoded, read-only file
	for name, node := range d.Nodes {
		direntType := fuse.DT_Unknown

		switch node.(type) {
		case *File:
			direntType = fuse.DT_File
		case *Dir:
			direntType = fuse.DT_Dir
		}

		entries = append(entries, fuse.Dirent{
			Name: name,
			Type: direntType,
		})
	}

	return entries, nil
}
