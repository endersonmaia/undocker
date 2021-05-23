package tartest

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	// Tarrer is an object that can tar itself to an archive
	Tarrer interface {
		Tar(*tar.Writer) error
	}

	// Tarball is a list of tarrables
	Tarball []Tarrer

	// Extractable is an empty interface for comparing extracted outputs in tests.
	// Using that just to avoid the ugly `interface{}`.
	Extractable interface{}

	// Dir is a tarrable directory representation
	Dir struct {
		Name string
		UID  int
	}

	// File is a tarrable file representation
	File struct {
		Name     string
		UID      int
		Contents *bytes.Buffer
	}

	// Hardlink is a representation of a hardlink
	Hardlink struct {
		Name string
		UID  int
	}
)

// Buffer returns a byte buffer
func (tb Tarball) Buffer() *bytes.Buffer {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, member := range tb {
		member.Tar(tw)
	}
	tw.Close()
	return &buf
}

// Tar tars the Dir
func (d Dir) Tar(tw *tar.Writer) error {
	hdr := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     d.Name,
		Mode:     0644,
		Uid:      d.UID,
	}
	return tw.WriteHeader(hdr)
}

// Tar tars the File
func (f File) Tar(tw *tar.Writer) error {
	var contents []byte
	if f.Contents != nil {
		contents = f.Contents.Bytes()
	}
	hdr := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     f.Name,
		Mode:     0644,
		Uid:      f.UID,
		Size:     int64(len(contents)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(contents); err != nil {
		return err
	}
	return nil
}

// Tar tars the Hardlink
func (h Hardlink) Tar(tw *tar.Writer) error {
	return tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeLink,
		Name:     h.Name,
		Mode:     0644,
		Uid:      h.UID,
	})
}

// Extract extracts a tarball to a slice of extractables
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
			elem = Dir{Name: hdr.Name, UID: hdr.Uid}
		case tar.TypeLink:
			elem = Hardlink{Name: hdr.Name}
		case tar.TypeReg:
			f := File{Name: hdr.Name, UID: hdr.Uid}
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
