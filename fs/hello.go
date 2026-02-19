package fs

import (
	"context"
	"fmt"
	"os"
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

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Inode: 2, Name: "hello.txt", Type: fuse.DT_File},
	}, nil
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

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Invalid usage. Use ./radFs <mntpoint>")
		return
	}

	mount_point := os.Args[1]

	//c is a fuse connection to dev/fuse
	c, err := fuse.Mount(mount_point)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer c.Close() //delay execution of Close

	err = fs.Serve(c, FS{})
	if err != nil {
		fmt.Println(err)
	} //starts listening for FS reqs

}
