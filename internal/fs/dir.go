package fs

import (
	"context"
	"log/slog"
	"os"
	"syscall"
	"time"

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
	a.Atime = d.atime
	a.Mtime = d.mtime
	a.Ctime = d.ctime
	a.Uid = d.uid
	a.Gid = d.gid

	return nil
}

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if req.Valid.Uid() {

		//if caller is not root and caller is trying to chown to uid that is not itself
		if req.Header.Uid != 0 && req.Uid != req.Header.Uid { // to get the uid of the process making the req -> checking the caller
			return syscall.EPERM
		}

		d.uid = req.Uid
		d.ctime = time.Now()
	}

	if req.Valid.Atime() {
		d.atime = req.Atime
	}
	if req.Valid.Mtime() {
		d.mtime = req.Mtime
	}
	d.ctime = time.Now()

	resp.Attr.Inode = d.inode
	resp.Attr.Mode = os.ModeDir | 0o755

	resp.Attr.Atime = d.atime
	resp.Attr.Mtime = d.mtime
	resp.Attr.Ctime = d.ctime
	resp.Attr.Uid = d.uid
	resp.Attr.Gid = d.gid

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

	d.atime = time.Now()

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

	d.atime = time.Now()

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

	newDir := &Dir{
		inode: nextInode(),
		Nodes: make(map[string]fs.Node),
		fs:    d.fs,
		atime: time.Now(),
		ctime: time.Now(),
		mtime: time.Now(),
		uid: req.Uid,
		gid: req.Gid,
	}
	d.Nodes[req.Name] = newDir

	d.mtime = time.Now()
	d.ctime = time.Now()

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

	f := &File{
		inode:  nextInode(),
		blocks: [][]byte{},
		size:   0,
		mode:   uint32(req.Mode),
		atime:  time.Now(),
		ctime:  time.Now(),
		mtime:  time.Now(),
		uid: req.Uid,
		gid: req.Gid,
	}

	if _, exists := d.Nodes[req.Name]; exists { // checking for dupes
		return nil, nil, syscall.EEXIST
	}
	d.Nodes[req.Name] = f

	d.mtime = time.Now()
	d.ctime = time.Now()

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

	if dir, flag := d.Nodes[req.Name].(*Dir); flag {
		if len(dir.Nodes) > 0 {
			return syscall.ENOTEMPTY
		}
	}

	delete(d.Nodes, req.Name)

	d.mtime = time.Now()
	d.ctime = time.Now()

	return nil
}

func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	d.fs.DebugPrint(
		"RENAME",
		"from", req.OldName,
		"to", req.NewName,
	)

	newParent, ok := newDir.(*Dir)
	if !ok {
		return syscall.EINVAL
	}

	//same name in same dir so do nothing
	if d == newParent && req.OldName == req.NewName {
		return nil
	}

	//safe locking
	if d == newParent {
		d.mu.Lock()
		defer d.mu.Unlock()
	} else {
		d.mu.Lock()
		newParent.mu.Lock()
		defer d.mu.Unlock()
		defer newParent.mu.Unlock()
	}

	//checks if source exists
	node, exists := d.Nodes[req.OldName]
	if !exists {
		return syscall.ENOENT
	}

	
	// if destination exists → overwrite 
if existing, exists := newParent.Nodes[req.NewName]; exists {
    // if it's a directory, check if empty
    if dir, ok := existing.(*Dir); ok {
        if len(dir.Nodes) > 0 {
            return syscall.ENOTEMPTY
        }
    }
    delete(newParent.Nodes, req.NewName)
}

	//removes from old
	delete(d.Nodes, req.OldName)

	//adds to new
	newParent.Nodes[req.NewName] = node

	return nil
}

