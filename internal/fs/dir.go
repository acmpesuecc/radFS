package fs

import (
	"context"
	"log/slog"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func (f *FS) DebugPrint(msg string, v ...any) {
	if f.Debug {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		logger.Info(msg, v...)
	}
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = d.inode
	a.Mode = os.ModeDir | 0o755

	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.fs.DebugPrint("LOOKUP", "fetching", name)

	d.mu.Lock()
	defer d.mu.Unlock()

	node, ok := d.Nodes[name]

	if !ok {
		return nil, syscall.ENOENT
	}

	return node, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.fs.DebugPrint("READDIR", "inode", d.inode)

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
	d.fs.DebugPrint(
		"MKDIR",
		"ID", req.ID,
		"Creating directory", req.Name,
		"NodeID", req.Node,
		"With mode", req.Mode,
		"Request PID", req.Pid,
	)

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Nodes[req.Name]; exists {
		return nil, syscall.EEXIST
	}

	newDir := &Dir{inode: nextInode(), Nodes: make(map[string]fs.Node), fs: d.fs}
	d.Nodes[req.Name] = newDir

	return newDir, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	d.fs.DebugPrint(
		"CREATE",
		"ID", req.ID,
		"Creating file", req.Name,
		"NodeID", req.Node,
		"With mode", req.Mode,
		"Request PID", req.Pid,
		"Access mode", req.Flags,
	)

	d.mu.Lock()
	defer d.mu.Unlock()

	f := &File{inode: nextInode(), data: []byte{}, mode: uint32(req.Mode)}
	d.Nodes[req.Name] = f

	return f, f, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	d.fs.DebugPrint(
		"REMOVE",
		"ID", req.ID,
		"Is this a directory?", req.Dir,
		"Removing file/dir", req.Name,
		"NodeID", req.ID,
		"Request PID", req.Pid,
	)

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Nodes[req.Name]; !exists {
		return syscall.ENOENT
	}

	delete(d.Nodes, req.Name)

	return nil
}
