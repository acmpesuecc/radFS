package fs

import (
	"context"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "hello.txt" {
		return &File{}, nil
	}
	return nil, fuse.ENOENT
}

// File struct
type File struct{}

// File attributes
func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0o444 // read-only
	a.Size = uint64(len("Hello from radFS!\n"))
	return nil
}

func (File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) error {

	// If opened in write mode → deny
	if req.Flags.IsWriteOnly() || req.Flags.IsReadWrite() {
		return fuse.Errno(syscall.EACCES)
	}

	return nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {

	content := []byte("Hello from radFS!\n")

	// If offset is beyond file size → EOF
	if req.Offset >= int64(len(content)) {
		resp.Data = []byte{}
		return nil
	}

	// Calculate how much we can safely read
	end := req.Offset + int64(req.Size)
	if end > int64(len(content)) {
		end = int64(len(content))
	}

	resp.Data = content[req.Offset:end]

	return nil
}
