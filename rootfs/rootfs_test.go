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

type tarrable interface {
	tar(*testing.T, *tar.Writer)
}

type dir struct {
	name string
	uid  int
}

func (d dir) tar(t *testing.T, tw *tar.Writer) {
	t.Helper()
	hdr := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     d.name,
		Uid:      d.uid,
	}
	require.NoError(t, tw.WriteHeader(hdr))
}

type file struct {
	name     string
	uid      int
	contents []byte
}

func (f file) tar(t *testing.T, tw *tar.Writer) {
	t.Helper()
	hdr := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     f.name,
		Uid:      f.uid,
		Size:     int64(len(f.contents)),
	}
	require.NoError(t, tw.WriteHeader(hdr))
	_, err := tw.Write(f.contents)
	require.NoError(t, err)
}

type manifest []string

func (m manifest) tar(t *testing.T, tw *tar.Writer) {
	t.Helper()
	b, err := json.Marshal(dockerManifestJSON{{Layers: m}})
	require.NoError(t, err)
	file{name: "manifest.json", uid: 0, contents: b}.tar(t, tw)
}

type tarball []tarrable

func (tb tarball) bytes(t *testing.T) []byte {
	t.Helper()

	buf := bytes.Buffer{}
	tw := tar.NewWriter(&buf)
	for _, member := range tb {
		member.tar(t, tw)
	}
	require.NoError(t, tw.Close())
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

func TestRootFS(t *testing.T) {
	_layer0 := tarball{
		dir{name: "/", uid: 0},
		file{name: "/file", uid: 1, contents: []byte("from 0")},
	}

	_layer1 := tarball{
		dir{name: "/", uid: 1},
		file{name: "/file", uid: 0, contents: []byte("from 1")},
	}

	_image := tarball{
		file{name: "layer1/layer.tar", contents: _layer1.bytes(t)},
		file{name: "layer0/layer.tar", contents: _layer0.bytes(t)},
		manifest{"layer0/layer0.tar", "layer1/layer.tar"},
	}

	tests := []struct {
		name  string
		image tarball
		want  []file
	}{
		{
			name:  "empty",
			image: tarball{manifest{}},
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
			in := bytes.NewReader(tt.image.bytes(t))
			out := bytes.Buffer{}

			require.NoError(t, RootFS(in, &out))
			got := extract(t, &out)
			assert.Equal(t, tt.want, got)
		})
	}
}
