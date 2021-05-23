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

func TestRootFS(t *testing.T) {
	layer0 := tarball{
		dir{name: "/", uid: 0},
		file{name: "/file", uid: 0, contents: []byte("from 0")},
	}

	layer1 := tarball{
		file{name: "/file", uid: 1, contents: []byte("from 1")},
	}

	layer2 := tarball{
		dir{name: "/", uid: 2},
	}

	tests := []struct {
		name    string
		image   tarball
		want    []extractable
		wantErr string
	}{
		{
			name:  "empty tarball",
			image: tarball{manifest{}},
			want:  []extractable{},
		},
		{
			name:    "missing layer",
			image:   tarball{manifest{"layer0/layer.tar"}},
			wantErr: "bad or missing manifest.json",
		},
		{
			name: "basic file overwrite, layer order mixed",
			image: tarball{
				file{name: "layer1/layer.tar", contents: layer1.bytes(t)},
				file{name: "layer0/layer.tar", contents: layer0.bytes(t)},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				dir{name: "/", uid: 0},
				file{name: "/file", uid: 1, contents: []byte("from 1")},
			},
		},
		{
			name: "directory overwrite retains original dir",
			image: tarball{
				file{name: "layer2/layer.tar", contents: layer2.bytes(t)},
				file{name: "layer0/layer.tar", contents: layer0.bytes(t)},
				file{name: "layer1/layer.tar", contents: layer1.bytes(t)},
				manifest{"layer0/layer.tar", "layer1/layer.tar", "layer2/layer.tar"},
			},
			want: []extractable{
				dir{name: "/", uid: 0},
				file{name: "/file", uid: 1, contents: []byte("from 1")},
				dir{name: "/", uid: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewReader(tt.image.bytes(t))
			out := bytes.Buffer{}

			err := RootFS(in, &out)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			got := extract(t, &out)
			assert.Equal(t, got, tt.want)
		})
	}
}

// Helpers

type tarrable interface {
	tar(*testing.T, *tar.Writer)
}

// extractable is an empty interface for comparing extracted outputs in tests.
// Using that just to avoid the ugly `interface{}`.
type extractable interface{}

type dir struct {
	name string
	uid  int
}

func (d dir) tar(t *testing.T, tw *tar.Writer) {
	t.Helper()
	hdr := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     d.name,
		Mode:     0644,
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
		Mode:     0644,
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

func extract(t *testing.T, f io.Reader) []extractable {
	t.Helper()
	ret := []extractable{}
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		var elem extractable
		switch hdr.Typeflag {
		case tar.TypeDir:
			elem = dir{name: hdr.Name, uid: hdr.Uid}
		case tar.TypeReg:
			buf := bytes.Buffer{}
			io.Copy(&buf, tr)
			elem = file{name: hdr.Name, uid: hdr.Uid, contents: buf.Bytes()}
		}
		ret = append(ret, elem)
	}
	return ret
}
