package tartest

import (
	"bytes"
	"compress/gzip"
	"io"
	"reflect"
	"testing"
)

func TestTarball(t *testing.T) {
	bk := Tarball{File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("hello")}}
	img := Tarball{
		File{Name: "backup.tar", Contents: bk.Buffer()},
		File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("bye")},
		Dir{Name: "bin"},
		Hardlink{Name: "entrypoint2"},
	}

	got := Extract(t, img.Buffer())
	want := []Extractable{
		File{Name: "backup.tar", Contents: bk.Buffer()},
		File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("bye")},
		Dir{Name: "bin"},
		Hardlink{Name: "entrypoint2"},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("tarball mismatch. want: %+v, got: %+v", want, got)
	}

}

func TestGzip(t *testing.T) {
	tb := Tarball{File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("hello")}}
	tbuf := tb.Buffer()

	tgz, err := gzip.NewReader(tb.Gzip())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var uncompressed bytes.Buffer
	if _, err := io.Copy(&uncompressed, tgz); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(tbuf.Bytes(), uncompressed.Bytes()) {
		t.Errorf("tbuf and uncompressed bytes mismatch")
	}
}
