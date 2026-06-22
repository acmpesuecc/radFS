package fs

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"bazil.org/fuse/fs"
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

	content := []byte("Hello from radFS!\n")
	blocks := [][]byte{}
	for i := 0; i < len(content); i += blockSize {
		end := min(i+blockSize, len(content))
		block := make([]byte, blockSize)
		copy(block, content[i:end])
		blocks = append(blocks, block)
	}

	root := &Dir{
		inode: 1,
		Nodes: map[string]fs.Node{

			"hello.txt": &File{
				inode:  nextInode(),
				blocks: blocks,
				size:   uint64(len(content)),
				mode:   0o666,
				atime:  time.Now(),
				mtime:  time.Now(),
				ctime:  time.Now(),
				uid:    uint32(os.Getuid()),
				gid:    uint32(os.Getgid()),
			},
		},
		fs:    f,
		atime: time.Now(),
		mtime: time.Now(),
		ctime: time.Now(),
		uid:    uint32(os.Getuid()),
		gid:    uint32(os.Getgid()),
	}

	return root, nil
}

type File struct {
	mu     sync.Mutex
	inode  uint64
	blocks [][]byte
	size   uint64
	mode   uint32
	atime  time.Time // read
	mtime  time.Time // write | truncate
	ctime  time.Time // metadata (setattr)
	uid    uint32
	gid    uint32
}

type Dir struct {
	mu    sync.Mutex
	inode uint64
	Nodes map[string]fs.Node
	fs    *FS
	atime time.Time
	mtime time.Time
	ctime time.Time
	uid   uint32
	gid   uint32
}
