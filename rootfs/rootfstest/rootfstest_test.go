package rootfstest

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarball(t *testing.T) {
	bk := Tarball{File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("hello")}}
	img := Tarball{
		File{Name: "backup.tar", Contents: bk},
		File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("bye")},
		Dir{Name: "bin"},
		Hardlink{Name: "entrypoint2"},
	}

	got := Extract(t, bytes.NewBuffer(img.Bytes()))
	want := []Extractable{
		File{Name: "backup.tar", Contents: bytes.NewBuffer(bk.Bytes())},
		File{Name: "entrypoint.sh", Contents: bytes.NewBufferString("bye")},
		Dir{Name: "bin"},
		Hardlink{Name: "entrypoint2"},
	}

	assert.Equal(t, want, got)
}
