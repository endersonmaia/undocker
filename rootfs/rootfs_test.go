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
		file{name: "/file", uid: 0, contents: bytes.NewBufferString("from 0")},
	}

	layer1 := tarball{
		file{name: "/file", uid: 1, contents: bytes.NewBufferString("from 1")},
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
				file{name: "layer1/layer.tar", contents: layer1},
				file{name: "layer0/layer.tar", contents: layer0},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				dir{name: "/", uid: 0},
				file{name: "/file", uid: 1, contents: bytes.NewBufferString("from 1")},
			},
		},
		{
			name: "directory overwrite retains original dir",
			image: tarball{
				file{name: "layer2/layer.tar", contents: layer2},
				file{name: "layer0/layer.tar", contents: layer0},
				file{name: "layer1/layer.tar", contents: layer1},
				manifest{"layer0/layer.tar", "layer1/layer.tar", "layer2/layer.tar"},
			},
			want: []extractable{
				dir{name: "/", uid: 0},
				file{name: "/file", uid: 1, contents: bytes.NewBufferString("from 1")},
				dir{name: "/", uid: 2},
			},
		},
		{
			name: "simple whiteout",
			image: tarball{
				file{name: "layer0/layer.tar", contents: tarball{
					file{name: "filea"},
					file{name: "fileb"},
					dir{name: "dira"},
					dir{name: "dirb"},
				}},
				file{name: "layer1/layer.tar", contents: tarball{
					hardlink{name: ".wh.filea"},
					hardlink{name: ".wh.dira"},
				}},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				file{name: "fileb"},
				dir{name: "dirb"},
			},
		},
		{
			name: "whiteout with override",
			image: tarball{
				file{name: "layer0/layer.tar", contents: tarball{
					file{name: "filea", contents: bytes.NewBufferString("from 0")},
				}},
				file{name: "layer1/layer.tar", contents: tarball{
					hardlink{name: ".wh.filea"},
				}},
				file{name: "layer2/layer.tar", contents: tarball{
					file{name: "filea", contents: bytes.NewBufferString("from 3")},
				}},
				manifest{
					"layer0/layer.tar",
					"layer1/layer.tar",
					"layer2/layer.tar",
				},
			},
			want: []extractable{
				file{name: "filea", contents: bytes.NewBufferString("from 3")},
			},
		},
		{
			name: "files and directories do not whiteout",
			image: tarball{
				file{name: "layer0/layer.tar", contents: tarball{
					dir{name: "dir"},
					file{name: "file"},
				}},
				file{name: "layer1/layer.tar", contents: tarball{
					dir{name: ".wh.dir"},
					file{name: ".wh.file"},
				}},
			},
			want: []extractable{
				dir{name: "dir"},
				dir{name: ".wh.dir"},
				file{name: "file"},
				file{name: ".wh.file"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewReader(tt.image.Bytes())
			out := bytes.Buffer{}

			err := RootFS(in, &out)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			got := extract(t, &out)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helpers

type tarrer interface {
	tar(*tar.Writer)
}

type byter interface {
	Bytes() []byte
}

type tarball []tarrer

func (tb tarball) Bytes() []byte {
	buf := bytes.Buffer{}
	tw := tar.NewWriter(&buf)
	for _, member := range tb {
		member.tar(tw)
	}
	tw.Close()
	return buf.Bytes()
}

// extractable is an empty interface for comparing extracted outputs in tests.
// Using that just to avoid the ugly `interface{}`.
type extractable interface{}

type dir struct {
	name string
	uid  int
}

func (d dir) tar(tw *tar.Writer) {
	hdr := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     d.name,
		Mode:     0644,
		Uid:      d.uid,
	}
	tw.WriteHeader(hdr)
}

type file struct {
	name     string
	uid      int
	contents byter
}

func (f file) tar(tw *tar.Writer) {
	var contentbytes []byte
	if f.contents != nil {
		contentbytes = f.contents.Bytes()
	}
	hdr := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     f.name,
		Mode:     0644,
		Uid:      f.uid,
		Size:     int64(len(contentbytes)),
	}
	tw.WriteHeader(hdr)
	tw.Write(contentbytes)
}

type manifest []string

func (m manifest) tar(tw *tar.Writer) {
	b, err := json.Marshal(dockerManifestJSON{{Layers: m}})
	if err != nil {
		panic("testerr")
	}
	file{
		name:     "manifest.json",
		uid:      0,
		contents: bytes.NewBuffer(b),
	}.tar(tw)
}

type hardlink struct {
	name string
	uid  int
}

func (h hardlink) tar(tw *tar.Writer) {
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeLink,
		Name:     h.name,
		Mode:     0644,
		Uid:      h.uid,
	})
}

func extract(t *testing.T, r io.Reader) []extractable {
	t.Helper()
	ret := []extractable{}
	tr := tar.NewReader(r)
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
			f := file{name: hdr.Name, uid: hdr.Uid}
			if hdr.Size > 0 {
				var buf bytes.Buffer
				io.Copy(&buf, tr)
				f.contents = &buf
			}
			elem = f
		}
		ret = append(ret, elem)
	}
	return ret
}
