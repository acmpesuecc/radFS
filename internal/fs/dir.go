package fs

import (
	"context"
	"log/slog"
	"os"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/acmpesuecc/radFS/internal/art"
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

	return nil
}

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	d.mu.Lock()
	defer d.mu.Unlock()

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

	return nil

}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.fs.DebugPrint("LOOKUP", "fetching", name)

	d.mu.RLock()
	defer d.mu.RUnlock()

	v, ok := d.tree.Search([]byte(name))

	if !ok {
		return nil, syscall.ENOENT
	}
	d.atime = time.Now()

	return v.(fs.Node), nil

}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.fs.DebugPrint("READDIR", "inode", d.inode)

	d.mu.RLock()
	defer d.mu.RUnlock()

	var entries []fuse.Dirent
	d.tree.ForEach(func(b []byte, i interface{}) { //traverses tree and appends the dirent to entries
		name := string(b)
		var dtype fuse.DirentType
		switch i.(type) {
		case *File:
			dtype = fuse.DT_File
		case *Dir:
			dtype = fuse.DT_Dir

		}
		entries = append(entries, fuse.Dirent{Name: name, Type: dtype})
	})

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

	if _, exists := d.tree.Search([]byte(req.Name)); exists {
		return nil, syscall.EEXIST
	}

	newDir := &Dir{
		inode: nextInode(),
		tree:  art.New(),
		fs:    d.fs,
		atime: time.Now(),
		mtime: time.Now(),
		ctime: time.Now(),
	}
	d.tree.Insert([]byte(req.Name), newDir)

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

	if _, exist := d.tree.Search([]byte(req.Name)); exist {
		return nil, nil, syscall.EEXIST

	}

	f := &File{inode: nextInode(), data: []byte{}, mode: uint32(req.Mode), atime: time.Now(),
		ctime: time.Now(),
		mtime: time.Now()}

	d.tree.Insert([]byte(req.Name), f)

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

	v, exist := d.tree.Search([]byte(req.Name))

	if !exist {
		return syscall.ENOENT
	}

	if dir, ok := v.(*Dir); ok {
		dir.mu.RLock() // we are reading another dir with Empty() , multiple processes may read
		defer dir.mu.RUnlock()
		if !dir.tree.Empty() {
			return syscall.ENOTEMPTY
		}
	}

	d.tree.Delete([]byte(req.Name))

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
	node, exists := d.tree.Search([]byte(req.OldName))
	if !exists {
		return syscall.ENOENT
	}

	// if destination exists → overwrite
	if existing, exists := newParent.tree.Search([]byte(req.NewName)); exists {
		// if it's a directory, check if empty
		if dir, ok := existing.(*Dir); ok {
			dir.mu.RLock()
			defer dir.mu.RUnlock()
			if !dir.tree.Empty() {
				return syscall.ENOTEMPTY
			}
		}
		newParent.tree.Delete([]byte(req.NewName))
	}

	//removes from old

	d.tree.Delete([]byte(req.OldName))

	//adds to new

	newParent.tree.Insert([]byte(req.NewName), node)
	return nil
}
