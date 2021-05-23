package tartest

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, want, got)
}

func TestGzip(t *testing.T) {
	tb := Tarball{File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("hello")}}
	tbuf := tb.Buffer()

	tgz, err := gzip.NewReader(tb.Gzip())
	require.NoError(t, err)
	var uncompressed bytes.Buffer
	_, err = io.Copy(&uncompressed, tgz)
	require.NoError(t, err)
	assert.Equal(t, tbuf.Bytes(), uncompressed.Bytes())

}
