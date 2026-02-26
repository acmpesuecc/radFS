package fs

import (
	"context"
	"log"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

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
	d.Nodes[req.Name] = new_dir

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
