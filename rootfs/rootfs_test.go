package rootfs

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"git.jakstys.lt/motiejus/undocker/rootfs/internal/tartest"
)

type (
	file        = tartest.File
	dir         = tartest.Dir
	hardlink    = tartest.Hardlink
	extractable = tartest.Extractable
	tarball     = tartest.Tarball
)

func TestRootFS(t *testing.T) {
	layer0 := tarball{
		dir{Name: "/", UID: 0},
		file{Name: "/file", UID: 0, Contents: bytes.NewBufferString("from 0")},
	}

	layer1 := tarball{
		file{Name: "/file", UID: 1, Contents: bytes.NewBufferString("from 1")},
	}

	layer2 := tarball{
		dir{Name: "/", UID: 2},
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
			name:    "no manifest",
			image:   tarball{},
			wantErr: "empty or missing manifest",
		},
		{
			name:    "missing layer",
			image:   tarball{manifest{"blobs/layer0/layer"}},
			wantErr: "blobs/layer0/layer defined in manifest, missing in tarball",
		},
		{
			name: "basic file overwrite, layer order mixed",
			image: tarball{
				file{Name: "blobs/layer1/layer", Contents: layer1.Buffer()},
				file{Name: "blobs/layer0/layer", Contents: layer0.Buffer()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer"},
			},
			want: []extractable{
				dir{Name: "/", UID: 0},
				file{Name: "/file", UID: 1, Contents: bytes.NewBufferString("from 1")},
			},
		},
		{
			name: "overwrite file with hardlink",
			image: tarball{
				file{Name: "blobs/layer0/layer", Contents: tarball{
					file{Name: "a"},
				}.Buffer()},
				file{Name: "blobs/layer1/layer", Contents: tarball{
					hardlink{Name: "a"},
				}.Buffer()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer"},
			},
			want: []extractable{
				hardlink{Name: "a"},
			},
		},
		{
			name: "directory overwrite retains original dir",
			image: tarball{
				file{Name: "blobs/layer2/layer", Contents: layer2.Buffer()},
				file{Name: "blobs/layer0/layer", Contents: layer0.Buffer()},
				file{Name: "blobs/layer1/layer", Contents: layer1.Buffer()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer", "blobs/layer2/layer"},
			},
			want: []extractable{
				dir{Name: "/", UID: 0},
				file{Name: "/file", UID: 1, Contents: bytes.NewBufferString("from 1")},
				dir{Name: "/", UID: 2},
			},
		},
		{
			name: "simple whiteout",
			image: tarball{
				file{Name: "blobs/layer0/layer", Contents: tarball{
					file{Name: "filea"},
					file{Name: "fileb"},
					dir{Name: "dira"},
					dir{Name: "dirb"},
				}.Buffer()},
				file{Name: "blobs/layer1/layer", Contents: tarball{
					hardlink{Name: ".wh.filea"},
					hardlink{Name: ".wh.dira"},
				}.Buffer()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer"},
			},
			want: []extractable{
				file{Name: "fileb"},
				dir{Name: "dirb"},
			},
		},
		{
			name: "whiteout with override",
			image: tarball{
				file{Name: "blobs/layer0/layer", Contents: tarball{
					file{Name: "file", Contents: bytes.NewBufferString("from 0")},
				}.Buffer()},
				file{Name: "blobs/layer1/layer", Contents: tarball{
					hardlink{Name: ".wh.file"},
				}.Buffer()},
				file{Name: "blobs/layer2/layer", Contents: tarball{
					file{Name: "file", Contents: bytes.NewBufferString("from 3")},
				}.Buffer()},
				manifest{
					"blobs/layer0/layer",
					"blobs/layer1/layer",
					"blobs/layer2/layer",
				},
			},
			want: []extractable{
				file{Name: "file", Contents: bytes.NewBufferString("from 3")},
			},
		},
		{
			name: "directories do not whiteout",
			image: tarball{
				file{Name: "blobs/layer0/layer", Contents: tarball{
					dir{Name: "dir"},
				}.Buffer()},
				file{Name: "blobs/layer1/layer", Contents: tarball{
					dir{Name: ".wh.dir"},
				}.Buffer()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer"},
			},
			want: []extractable{
				dir{Name: "dir"},
				dir{Name: ".wh.dir"},
			},
		},
		{
			name: "simple readdir whiteout",
			image: tarball{
				file{Name: "blobs/layer0/layer", Contents: tarball{
					dir{Name: "a"},
					file{Name: "a/filea"},
				}.Buffer()},
				file{Name: "blobs/layer1/layer", Contents: tarball{
					dir{Name: "a"},
					file{Name: "a/fileb"},
					hardlink{Name: "a/.wh..wh..opq"},
				}.Buffer()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer"},
			},
			want: []extractable{
				dir{Name: "a"},
				file{Name: "a/fileb"},
			},
		},
		{
			name: "archived layer",
			image: tarball{
				file{Name: "blobs/layer1/layer", Contents: layer1.Gzip()},
				file{Name: "blobs/layer0/layer", Contents: layer0.Gzip()},
				manifest{"blobs/layer0/layer", "blobs/layer1/layer"},
			},
			want: []extractable{
				dir{Name: "/", UID: 0},
				file{Name: "/file", UID: 1, Contents: bytes.NewBufferString("from 1")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewReader(tt.image.Buffer().Bytes())
			out := bytes.Buffer{}

			err := Flatten(in, &out)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErr != err.Error() {
					t.Errorf("want != got: %s != %s", tt.wantErr, err.Error())
				}
				return
			}
			outb := out.Bytes()
			if err != nil {
				t.Fatal("expected error, got nil")
			}
			got := tartest.Extract(t, bytes.NewReader(outb))
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want != got: %v != %v", tt.want, got)
			}
		})
	}
}

// Helpers
type manifest []string

func (m manifest) Tar(tw *tar.Writer) error {
	b, err := json.Marshal(dockerManifestJSON{{Layers: m}})
	if err != nil {
		return err
	}
	return file{
		Name:     "manifest.json",
		Contents: bytes.NewBuffer(b),
	}.Tar(tw)
}
