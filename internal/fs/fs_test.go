package fs

import (
	"context"
	"os"
	"testing"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func newTestFS() *FS       { return New(false) }
func ctx() context.Context { return context.Background() }

func rootDir(t *testing.T) *Dir {
	t.Helper()
	f := newTestFS()
	node, err := f.Root()
	if err != nil {
		t.Fatalf("Root() error: %v", err)
	}
	d, ok := node.(*Dir)
	if !ok {
		t.Fatalf("Root() did not return *Dir")
	}
	return d
}

func TestRoot_ReturnsDir(t *testing.T) {
	d := rootDir(t)
	if d == nil {
		t.Fatal("expected non-nil *Dir")
	}
}

func TestRoot_HasHelloTxt(t *testing.T) {
	d := rootDir(t)
	if _, ok := d.tree.Search([]byte("hello.txt")); !ok {
		t.Error("root dir missing hello.txt")
	}
}

func TestLookup_ExistingFile(t *testing.T) {
	d := rootDir(t)
	node, err := d.Lookup(ctx(), "hello.txt")
	if err != nil {
		t.Fatalf("Lookup existing file: %v", err)
	}
	if node == nil {
		t.Fatal("Lookup returned nil node")
	}
}

func TestLookup_MissingFile(t *testing.T) {
	d := rootDir(t)
	_, err := d.Lookup(ctx(), "no-such-file.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestCreate_NewFile(t *testing.T) {
	d := rootDir(t)
	req := &fuse.CreateRequest{
		Name:  "new.txt",
		Flags: fuse.OpenReadWrite,
		Mode:  0o666,
	}
	resp := &fuse.CreateResponse{}

	node, handle, err := d.Create(ctx(), req, resp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if node == nil {
		t.Fatal("Create returned nil node")
	}
	if handle == nil {
		t.Fatal("Create returned nil handle")
	}

	if _, ok := d.tree.Search([]byte("new.txt")); !ok {
		t.Error("new.txt not found in dir after Create")
	}
}

func TestCreate_DuplicateFile(t *testing.T) {
	d := rootDir(t)
	req := &fuse.CreateRequest{Name: "dup.txt", Mode: 0o666}
	resp := &fuse.CreateResponse{}

	if _, _, err := d.Create(ctx(), req, resp); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, _, _ = d.Create(ctx(), req, resp)
}

func TestMkdir_NewDir(t *testing.T) {
	d := rootDir(t)
	req := &fuse.MkdirRequest{Name: "subdir", Mode: os.ModeDir | 0o755}

	node, err := d.Mkdir(ctx(), req)
	if err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if node == nil {
		t.Fatal("Mkdir returned nil node")
	}
	if _, ok := d.tree.Search([]byte("subdir")); !ok {
		t.Error("subdir not found in dir after Mkdir")
	}
}

func TestMkdir_NestedLookup(t *testing.T) {
	d := rootDir(t)
	req := &fuse.MkdirRequest{Name: "inner", Mode: os.ModeDir | 0o755}
	if _, err := d.Mkdir(ctx(), req); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	node, err := d.Lookup(ctx(), "inner")
	if err != nil {
		t.Fatalf("Lookup after Mkdir: %v", err)
	}
	if _, ok := node.(*Dir); !ok {
		t.Error("Lookup of mkdir result is not *Dir")
	}
}

func TestWrite_BasicContent(t *testing.T) {
	f := &File{inode: nextInode(), mode: 0o666}
	req := &fuse.WriteRequest{Data: []byte("hello world"), Offset: 0}
	resp := &fuse.WriteResponse{}

	if err := f.Write(ctx(), req, resp); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if resp.Size != len(req.Data) {
		t.Errorf("Write size = %d, want %d", resp.Size, len(req.Data))
	}
}

func TestRead_AfterWrite(t *testing.T) {
	content := []byte("round-trip content")
	f := &File{inode: nextInode(), data: content, mode: 0o666}

	req := &fuse.ReadRequest{Offset: 0, Size: len(content)}
	resp := &fuse.ReadResponse{}

	if err := f.Read(ctx(), req, resp); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(resp.Data) != string(content) {
		t.Errorf("Read = %q, want %q", resp.Data, content)
	}
}

func TestRead_PartialOffset(t *testing.T) {
	f := &File{inode: nextInode(), data: []byte("0123456789"), mode: 0o666}
	req := &fuse.ReadRequest{Offset: 4, Size: 4}
	resp := &fuse.ReadResponse{}

	if err := f.Read(ctx(), req, resp); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(resp.Data) != "4567" {
		t.Errorf("Read partial = %q, want %q", resp.Data, "4567")
	}
}

func TestRead_BeyondEOF(t *testing.T) {
	f := &File{inode: nextInode(), data: []byte("short"), mode: 0o666}
	req := &fuse.ReadRequest{Offset: 100, Size: 10}
	resp := &fuse.ReadResponse{}

	if err := f.Read(ctx(), req, resp); err != nil {
		t.Fatalf("Read beyond EOF should not error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 bytes beyond EOF, got %d", len(resp.Data))
	}
}

func TestRemove_ExistingFile(t *testing.T) {
	d := rootDir(t)
	req := &fuse.RemoveRequest{Name: "hello.txt", Dir: false}

	if err := d.Remove(ctx(), req); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, ok := d.tree.Search([]byte("hello.txt")); ok {
		t.Error("hello.txt still present after Remove")
	}
}

func TestRemove_MissingFile(t *testing.T) {
	d := rootDir(t)
	req := &fuse.RemoveRequest{Name: "ghost.txt", Dir: false}

	err := d.Remove(ctx(), req)
	if err == nil {
		t.Fatal("expected error removing non-existent file")
	}
}

func TestRemove_ThenLookupFails(t *testing.T) {
	d := rootDir(t)
	_ = d.Remove(ctx(), &fuse.RemoveRequest{Name: "hello.txt"})

	_, err := d.Lookup(ctx(), "hello.txt")
	if err == nil {
		t.Fatal("Lookup after Remove should return error")
	}
}

func TestFileAttr(t *testing.T) {
	f := &File{inode: 42, data: []byte("test"), mode: 0o644}
	var a fuse.Attr
	if err := f.Attr(ctx(), &a); err != nil {
		t.Fatalf("File.Attr: %v", err)
	}
	if a.Inode != 42 {
		t.Errorf("inode = %d, want 42", a.Inode)
	}
	if a.Size != 4 {
		t.Errorf("size = %d, want 4", a.Size)
	}
}

func TestDirAttr(t *testing.T) {
	d := rootDir(t)
	var a fuse.Attr
	if err := d.Attr(ctx(), &a); err != nil {
		t.Fatalf("Dir.Attr: %v", err)
	}
	if a.Inode != 1 {
		t.Errorf("root inode = %d, want 1", a.Inode)
	}
	if a.Mode&os.ModeDir == 0 {
		t.Error("Dir.Attr mode missing ModeDir bit")
	}
}

func TestReadDirAll_ContainsHello(t *testing.T) {
	d := rootDir(t)
	entries, err := d.ReadDirAll(ctx())
	if err != nil {
		t.Fatalf("ReadDirAll: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Name == "hello.txt" {
			found = true
		}
	}
	if !found {
		t.Error("ReadDirAll missing hello.txt entry")
	}
}

var _ fs.Node = (*File)(nil)
var _ fs.Node = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)
var _ fs.HandleReader = (*File)(nil)
