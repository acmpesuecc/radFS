package fs

import (
	"context"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Dir struct {
	Nodes map[string]fs.Node
}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1                 //inode 1 cuz root
	a.Mode = os.ModeDir | 0o755 //octal perms for rwx r-x r-x

	return nil
}

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

// dirs dont need handle as no I/O for them
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {

	name := req.Name

	if debug {
		log.Printf("New Dir Called at: %s", name)
	}

	_, exists := d.Nodes[name] //value not needed => make it _

	//dupes
	if exists {
		return nil, syscall.EEXIST
	}

	//create dir
	new_dir := &Dir{
		Nodes: make(map[string]fs.Node),
	}
	d.Nodes[req.Name] = new_dir //adding to parent dir

	return new_dir, nil

}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {

	file_name := req.Name

	new_file := &File{
		Name: file_name,
		Data: []byte{},
	}

	d.Nodes[file_name] = new_file

	if debug {
		log.Printf("New file Create called at: %s", file_name)
	}

	return new_file, new_file, nil

}
