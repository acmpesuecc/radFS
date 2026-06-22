package fs

import (
	"context"
	"os"
	"time"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

const blockSize = 4096

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a.Inode = f.inode
	a.Mode = os.FileMode(f.mode)
	a.Size = uint64(f.size) //using size instead of len because its [][]byte
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

	// checking if reading is starting from offset that actually exists
	if req.Offset >= int64(f.size) {
		resp.Data = []byte{}
		return nil
	}

	// so that we dont read past file size
	end := req.Offset + int64(req.Size)
	if end > int64(f.size) {
		end = int64(f.size)
	}

	var result []byte // final fully read file as a single array that we will reeturn
	offset := req.Offset

	for offset < end {
		blockIndex := offset / blockSize
		blockOffset := offset % blockSize

		// min of how much space is left on currect block and how much data is left to read -> basically safety guard
		toRead := min(int64(blockSize)-blockOffset, end-offset)

		result = append(result, f.blocks[blockIndex][blockOffset:blockOffset+toRead]...) //expanding each byte individually

		offset += toRead
	}

	resp.Data = result
	f.atime = time.Now()

	f.atime = time.Now()

	return nil
}

// Write writes data to the file
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	offset := req.Offset
	data := req.Data
	end := req.Offset + int64(len(req.Data)) // till where we'll be writing

	// grow blocks if needed
	for end > int64(len(f.blocks))*blockSize {
		f.blocks = append(f.blocks, make([]byte, blockSize))
	}

	// updating file size
	if end > int64(f.size) {
		f.size = uint64(end)
	}

	written := 0

	for len(data) > 0 {
		blockIndex := offset / blockSize
		blockOffset := offset % blockSize

		toWrite := min(int64(blockSize)-blockOffset, int64(len(data)))

		n := copy(f.blocks[blockIndex][blockOffset:blockOffset+toWrite], data[:toWrite])

		data = data[n:]
		offset += int64(n)
		written += n

	}

	resp.Size = written
	f.mtime = time.Now()
	f.ctime = time.Now() // writing to file constitutes changes in certain fields of inode too

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

	if req.Valid.Uid() {

		//if caller is not root and caller is trying to chown to uid that is not itself
		if req.Header.Uid != 0 && req.Uid != req.Header.Uid{ // to get the uid of the process making the req -> checking the caller
			return syscall.EPERM
		}

		f.uid = req.Uid
		f.ctime = time.Now()
	}

	if req.Valid.Gid() {
		f.gid = req.Gid
		f.ctime = time.Now()
	}

	if req.Valid.Size() { // mainly for truncate?

		newSize := req.Size
		needed := (newSize + uint64(blockSize) - 1) / uint64(blockSize) // ceil division to find how many blocks are required to store the new size

		if newSize < f.size { //shrinking operation

			//start from where we need to start
			f.blocks = f.blocks[:needed]

			if newSize > 0 {
				last_offset := newSize % uint64(blockSize)
				if last_offset != 0 {
					clear(f.blocks[needed-1][last_offset:]) // since we are shirnking/truncating we need to remove(clear) the stuff we dont need
				}
			}
		} else if newSize > f.size { // explanding operation
			for uint64(len(f.blocks)) < needed {
				f.blocks = append(f.blocks, make([]byte, blockSize))
			}
		}

		f.size = newSize
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
	resp.Attr.Size = f.size
	resp.Attr.Atime = f.atime
	resp.Attr.Mtime = f.mtime
	resp.Attr.Ctime = f.ctime
	resp.Attr.Uid = f.uid
	resp.Attr.Gid = f.gid

	return nil
}

// Flush is called when a file handle is closed
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return nil
}

func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	return nil
}
