package rootfs

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"testing"
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
	contents []byte
	uid      int
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

func RootFSTest(t *testing.T) {

}
