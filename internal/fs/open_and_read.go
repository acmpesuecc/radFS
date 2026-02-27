package fs

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a.Inode = f.inode
	a.Mode = os.FileMode(f.mode)
	a.Size = uint64(len(f.data))
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
	if end > int64(len(f.data)) {
		end = int64(len(f.data))
	}
	resp.Data = f.data[req.Offset:end]
	return nil
}

// Write writes data to the file
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	end := req.Offset + int64(len(req.Data))

	// Grow the buffer if needed
	if end > int64(len(f.data)) {
		newData := make([]byte, end)
		copy(newData, f.data)
		f.data = newData
	}

	copy(f.data[req.Offset:], req.Data)
	resp.Size = len(req.Data)
	return nil
}

// Setattr handles chmod, truncate, etc.
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Valid.Mode() {
		f.mode = uint32(req.Mode)
	}

	if req.Valid.Size() {
		if req.Size < uint64(len(f.data)) {
			f.data = f.data[:req.Size]
		} else {
			newData := make([]byte, req.Size)
			copy(newData, f.data)
			f.data = newData
		}
	}

	resp.Attr.Inode = f.inode
	resp.Attr.Mode = os.FileMode(f.mode)
	resp.Attr.Size = uint64(len(f.data))
	return nil
}

// Flush is called when a file handle is closed
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return nil
}
