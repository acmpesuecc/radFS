package fs

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"bazil.org/fuse/fs"
	"github.com/acmpesuecc/radFS/internal/art"
)

type FS struct {
	Debug bool
}

func New(debug bool) *FS {
	return &FS{Debug: debug}
}

var inodeCounter uint64 = 2

func nextInode() uint64 {
	return atomic.AddUint64(&inodeCounter, 1)
}

func (f *FS) Root() (fs.Node, error) {
	now := time.Now()
	uid := uint32(os.Getuid())
	gid := uint32(os.Getgid())

	root := &Dir{
		inode: 1,
		tree:  art.New(),
		fs:    f,
		atime: now,
		mtime: now,
		ctime: now,
		uid:   uid,
		gid:   gid,
	}

	hello := &File{
		inode: nextInode(),
		data:  []byte("Hello from radFS!\n"),
		mode:  0o666,
		atime: now,
		mtime: now,
		ctime: now,
		uid:   uid, //permissions implemnet based on userid
		gid:   gid, //permissions implement based on groupid
	}

	root.tree.Insert([]byte("hello.txt"), hello)

	return root, nil
}

type File struct {
	mu    sync.RWMutex
	inode uint64
	data  []byte
	mode  uint32
	atime time.Time // read
	mtime time.Time // write | truncate
	ctime time.Time // metadata (setattr)
	uid   uint32
	gid   uint32
}

type Dir struct {
	mu    sync.RWMutex
	inode uint64
	tree  *art.Tree
	fs    *FS
	atime time.Time
	mtime time.Time
	ctime time.Time
	uid   uint32
	gid   uint32
}
