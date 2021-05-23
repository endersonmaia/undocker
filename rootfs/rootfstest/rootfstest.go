package rootfstest

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	Tarrer interface {
		Tar(*tar.Writer)
	}

	Byter interface {
		Bytes() []byte
	}

	Tarball []Tarrer

	// Extractable is an empty interface for comparing extracted outputs in tests.
	// Using that just to avoid the ugly `interface{}`.
	Extractable interface{}

	Dir struct {
		Name string
		Uid  int
	}

	File struct {
		Name     string
		Uid      int
		Contents Byter
	}

	Manifest []string

	Hardlink struct {
		Name string
		Uid  int
	}

	dockerManifestJSON []struct {
		Layers []string `json:"Layers"`
	}
)

func (tb Tarball) Bytes() []byte {
	buf := bytes.Buffer{}
	tw := tar.NewWriter(&buf)
	for _, member := range tb {
		member.Tar(tw)
	}
	tw.Close()
	return buf.Bytes()
}

func (d Dir) Tar(tw *tar.Writer) {
	hdr := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     d.Name,
		Mode:     0644,
		Uid:      d.Uid,
	}
	tw.WriteHeader(hdr)
}

func (f File) Tar(tw *tar.Writer) {
	var contentbytes []byte
	if f.Contents != nil {
		contentbytes = f.Contents.Bytes()
	}
	hdr := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     f.Name,
		Mode:     0644,
		Uid:      f.Uid,
		Size:     int64(len(contentbytes)),
	}
	tw.WriteHeader(hdr)
	tw.Write(contentbytes)
}

func (m Manifest) Tar(tw *tar.Writer) {
	b, err := json.Marshal(dockerManifestJSON{{Layers: m}})
	if err != nil {
		panic("testerr")
	}
	File{
		Name:     "manifest.json",
		Uid:      0,
		Contents: bytes.NewBuffer(b),
	}.Tar(tw)
}

func (h Hardlink) Tar(tw *tar.Writer) {
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeLink,
		Name:     h.Name,
		Mode:     0644,
		Uid:      h.Uid,
	})
}

func Extract(t *testing.T, r io.Reader) []Extractable {
	t.Helper()
	ret := []Extractable{}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		var elem Extractable
		switch hdr.Typeflag {
		case tar.TypeDir:
			elem = Dir{Name: hdr.Name, Uid: hdr.Uid}
		case tar.TypeLink:
			elem = Hardlink{Name: hdr.Name}
		case tar.TypeReg:
			f := File{Name: hdr.Name, Uid: hdr.Uid}
			if hdr.Size > 0 {
				var buf bytes.Buffer
				io.Copy(&buf, tr)
				f.Contents = &buf
			}
			elem = f
		}
		ret = append(ret, elem)
	}
	return ret
}
