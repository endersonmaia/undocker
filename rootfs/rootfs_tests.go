package rootfs

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type layers []string

func (l layers) bytes() []byte {
	dockerManifest := dockerManifestJSON{{Layers: l}}
	b, err := json.Marshal(dockerManifest)
	if err != nil {
		panic("panic in a unit test")
	}
	return b
}

type tarrable interface {
	tar(*tar.Writer)
}

type file struct {
	name     string
	uid      int
	contents []byte
}

func (f file) tar(tw *tar.Writer) {
	hdr := &tar.Header{Typeflag: tar.TypeReg, Uid: f.uid}
	tw.WriteHeader(hdr)
	tw.Write(f.contents)
}

type dir struct {
	name string
	uid  int
}

func (f dir) tar(tw *tar.Writer) {
	hdr := &tar.Header{Typeflag: tar.TypeDir, Uid: f.uid}
	tw.WriteHeader(hdr)
}

type tarball []tarrable

func (t tarball) bytes() []byte {
	buf := bytes.Buffer{}
	tw := tar.NewWriter(&buf)
	for _, member := range t {
		member.tar(tw)
	}
	tw.Close()
	return buf.Bytes()
}

func extract(t *testing.T, tarball io.Reader) []file {
	t.Helper()
	var ret []file
	tr := tar.NewReader(tarball)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		elem := file{name: hdr.Name, uid: hdr.Uid}
		if hdr.Typeflag == tar.TypeReg {
			buf := bytes.Buffer{}
			io.Copy(&buf, tr)
			elem.contents = buf.Bytes()
		}
	}
	return ret
}

var (
	_layer0 = tarball{
		dir{name: "/", uid: 0},
		file{name: "/file", uid: 1, contents: []byte("from 0")},
	}

	_layer1 = tarball{
		dir{name: "/", uid: 1},
		file{name: "/file", uid: 0, contents: []byte("from 1")},
	}

	_image = tarball{
		file{name: "layer1/layer.tar", contents: _layer1.bytes()},
		file{name: "layer0/layer.tar", contents: _layer0.bytes()},
		file{
			name:     "manifest.json",
			contents: layers{"layer0/layer.tar", "layer1/layer0.tar"}.bytes(),
		},
	}
)

func TestRootFS(t *testing.T) {
	tests := []struct {
		name  string
		image tarball
		want  []file
	}{
		{
			name:  "empty",
			image: tarball{},
			want:  []file{},
		},
		{
			name:  "basic overwrite, 2 layers",
			image: _image,
			want: []file{
				{name: "/", uid: 1},
				{name: "file", uid: 0, contents: []byte("from 1")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewReader(tt.image.bytes())
			out := bytes.Buffer{}

			require.NoError(t, RootFS(in, &out))
			got := extract(t, &out)
			assert.Equal(t, tt.want, got)
		})
	}
}
