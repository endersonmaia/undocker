package rootfs

import (
	"bytes"
	"testing"

	"github.com/motiejus/code/undocker/rootfs/rootfstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	file        = rootfstest.File
	dir         = rootfstest.Dir
	hardlink    = rootfstest.Hardlink
	manifest    = rootfstest.Manifest
	extractable = rootfstest.Extractable
	tarball     = rootfstest.Tarball
)

var (
	extract = rootfstest.Extract
)

func TestRootFS(t *testing.T) {
	layer0 := tarball{
		dir{Name: "/", Uid: 0},
		file{Name: "/file", Uid: 0, Contents: bytes.NewBufferString("from 0")},
	}

	layer1 := tarball{
		file{Name: "/file", Uid: 1, Contents: bytes.NewBufferString("from 1")},
	}

	layer2 := tarball{
		dir{Name: "/", Uid: 2},
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
				file{Name: "layer1/layer.tar", Contents: layer1},
				file{Name: "layer0/layer.tar", Contents: layer0},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				dir{Name: "/", Uid: 0},
				file{Name: "/file", Uid: 1, Contents: bytes.NewBufferString("from 1")},
			},
		},
		{
			name: "overwrite file with hardlink",
			image: tarball{
				file{Name: "layer0/layer.tar", Contents: tarball{
					file{Name: "a"},
				}},
				file{Name: "layer1/layer.tar", Contents: tarball{
					hardlink{Name: "a"},
				}},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				hardlink{Name: "a"},
			},
		},
		{
			name: "directory overwrite retains original dir",
			image: tarball{
				file{Name: "layer2/layer.tar", Contents: layer2},
				file{Name: "layer0/layer.tar", Contents: layer0},
				file{Name: "layer1/layer.tar", Contents: layer1},
				manifest{"layer0/layer.tar", "layer1/layer.tar", "layer2/layer.tar"},
			},
			want: []extractable{
				dir{Name: "/", Uid: 0},
				file{Name: "/file", Uid: 1, Contents: bytes.NewBufferString("from 1")},
				dir{Name: "/", Uid: 2},
			},
		},
		{
			name: "simple whiteout",
			image: tarball{
				file{Name: "layer0/layer.tar", Contents: tarball{
					file{Name: "filea"},
					file{Name: "fileb"},
					dir{Name: "dira"},
					dir{Name: "dirb"},
				}},
				file{Name: "layer1/layer.tar", Contents: tarball{
					hardlink{Name: ".wh.filea"},
					hardlink{Name: ".wh.dira"},
				}},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				file{Name: "fileb"},
				dir{Name: "dirb"},
			},
		},
		{
			name: "whiteout with override",
			image: tarball{
				file{Name: "layer0/layer.tar", Contents: tarball{
					file{Name: "file", Contents: bytes.NewBufferString("from 0")},
				}},
				file{Name: "layer1/layer.tar", Contents: tarball{
					hardlink{Name: ".wh.file"},
				}},
				file{Name: "layer2/layer.tar", Contents: tarball{
					file{Name: "file", Contents: bytes.NewBufferString("from 3")},
				}},
				manifest{
					"layer0/layer.tar",
					"layer1/layer.tar",
					"layer2/layer.tar",
				},
			},
			want: []extractable{
				file{Name: "file", Contents: bytes.NewBufferString("from 3")},
			},
		},
		{
			name: "directories do not whiteout",
			image: tarball{
				file{Name: "layer0/layer.tar", Contents: tarball{
					dir{Name: "dir"},
				}},
				file{Name: "layer1/layer.tar", Contents: tarball{
					dir{Name: ".wh.dir"},
				}},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				dir{Name: "dir"},
				dir{Name: ".wh.dir"},
			},
		},
		{
			name: "simple readdir whiteout",
			image: tarball{
				file{Name: "layer0/layer.tar", Contents: tarball{
					dir{Name: "a"},
					file{Name: "a/filea"},
				}},
				file{Name: "layer1/layer.tar", Contents: tarball{
					dir{Name: "a"},
					file{Name: "a/fileb"},
					hardlink{Name: "a/.wh..wh..opq"},
				}},
				manifest{"layer0/layer.tar", "layer1/layer.tar"},
			},
			want: []extractable{
				dir{Name: "a"},
				file{Name: "a/fileb"},
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
