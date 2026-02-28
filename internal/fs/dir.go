package fs

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = d.inode
	a.Mode = os.ModeDir | 0o755
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	node, ok := d.Nodes[name]
	if !ok {
		return nil, fuse.ENOENT
	}
	return node, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var entries []fuse.Dirent
	for name, node := range d.Nodes {
		var dt fuse.DirentType
		switch node.(type) {
		case *Dir:
			dt = fuse.DT_Dir
		default:
			dt = fuse.DT_File
		}
		entries = append(entries, fuse.Dirent{Name: name, Type: dt})
	}
	return entries, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.Nodes[req.Name]; exists {
		return nil, fuse.EEXIST
	}
	newDir := &Dir{inode: nextInode(), Nodes: make(map[string]fs.Node)}
	d.Nodes[req.Name] = newDir
	return newDir, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	f := &File{inode: nextInode(), data: []byte{}, mode: uint32(req.Mode)}
	d.Nodes[req.Name] = f
	return f, f, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.Nodes[req.Name]; !exists {
		return fuse.ENOENT
	}
	delete(d.Nodes, req.Name)
	return nil
}
