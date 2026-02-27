package fs

import "os"

func (f *File) resize(size int) {
	if size < len(f.data) {
		f.data = f.data[:size]
		return
	}

	newBuf := make([]byte, size)
	copy(newBuf, f.data)
	f.data = newBuf
}

func (f *File) setMode(mode os.FileMode) {
	if mode == 0 {
		return
	}
	f.mode = mode
}
