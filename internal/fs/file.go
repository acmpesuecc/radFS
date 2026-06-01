package fs

import (
	"context"
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	a.Inode = f.inode
	a.Mode = os.FileMode(f.mode)
	a.Size = uint64(len(f.data))
	a.Atime = f.atime
	a.Mtime = f.mtime
	a.Ctime = f.ctime
	a.Uid = f.uid
	a.Gid = f.gid

	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return f, nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Offset >= int64(len(f.data)) {
		resp.Data = []byte{}

		return nil
	}

	end := req.Offset + int64(req.Size)
	end = min(end, int64(len(f.data)))
	resp.Data = f.data[req.Offset:end]

	f.atime = time.Now()

	return nil
}

// Write writes data to the file
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	end := req.Offset + int64(len(req.Data))

	// Grow the buffer if needed
	if end > int64(len(f.data)) {
		newData := make([]byte, int(end))
		copy(newData, f.data)
		f.data = newData
	}

	copy(f.data[req.Offset:], req.Data)
	resp.Size = len(req.Data)

	f.mtime = time.Now()
	f.ctime = time.Now() // writing to file constitutes changes in certain fields of inode too

	return nil
}

// Setattr handles chmod, truncate, etc.
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Valid.Mode() {
		f.mode = uint32(req.Mode)
		f.ctime = time.Now()
	}

	if req.Valid.Size() {
		if req.Size < uint64(len(f.data)) {
			f.data = f.data[:req.Size]
			f.ctime = time.Now() // cuz creating file here
		} else {
			newData := make([]byte, int(req.Size))
			copy(newData, f.data)
			f.data = newData
		}
		f.mtime = time.Now()
		f.ctime = time.Now()
	}

	if req.Valid.Atime() {
		f.atime = req.Atime
	}
	if req.Valid.Mtime() {
		f.mtime = req.Mtime
	}

	resp.Attr.Inode = f.inode
	resp.Attr.Mode = os.FileMode(f.mode)
	resp.Attr.Size = uint64(len(f.data))
	resp.Attr.Atime = f.atime
	resp.Attr.Mtime = f.mtime
	resp.Attr.Ctime = f.ctime

	return nil
}

// Flush is called when a file handle is closed
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return nil
}

func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	return nil
}
