package rootfs

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
)

const (
	_manifestJSON = "manifest.json"
	_layerSuffix  = "/layer.tar"
	_whReaddir    = ".wh..wh..opq"
	_whPrefix     = ".wh."
)

var (
	errBadManifest = errors.New("bad or missing manifest.json")
)

type dockerManifestJSON []struct {
	Layers []string `json:"Layers"`
}

// RootFS accepts a docker layer tarball and flattens it.
type RootFS struct {
	rd io.ReadSeeker
}

// New creates a new RootFS'er.
func New(rd io.ReadSeeker) *RootFS {
	return &RootFS{rd: rd}
}

// WriteTo writes a docker image to an open tarball.
func (r *RootFS) WriteTo(w io.Writer) (n int64, err error) {
	wr := &byteCounter{rw: w}
	tr := tar.NewReader(r.rd)
	tw := tar.NewWriter(wr)
	defer func() {
		err = multierr.Append(err, tw.Close())
		n = wr.n
	}()

	// layerOffsets maps a layer name (a9b123c0daa/layer.tar) to it's offset
	layerOffsets := map[string]int64{}

	// manifest is the docker manifest in the image
	var manifest dockerManifestJSON

	// get layer offsets and manifest.json
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		switch {
		case filepath.Clean(hdr.Name) == _manifestJSON:
			dec := json.NewDecoder(tr)
			if err := dec.Decode(&manifest); err != nil {
				return n, fmt.Errorf("decode %s: %w", _manifestJSON, err)
			}
		case strings.HasSuffix(hdr.Name, _layerSuffix):
			here, err := r.rd.Seek(0, io.SeekCurrent)
			if err != nil {
				return n, err
			}
			layerOffsets[hdr.Name] = here
		}
	}

	if len(manifest) == 0 || len(layerOffsets) != len(manifest[0].Layers) {
		return n, errBadManifest
	}

	// enumerate layers the way they would be laid down in the image
	layers := make([]int64, len(layerOffsets))
	for i, name := range manifest[0].Layers {
		layers[i] = layerOffsets[name]
	}

	// file2layer maps a filename to layer number (index in "layers")
	file2layer := map[string]int{}

	// whreaddir maps `wh..wh..opq` file to a layer; see doc.go
	whreaddir := map[string]int{}

	// wh maps a filename to a layer until which it should be ignored,
	// inclusively; see doc.go
	wh := map[string]int{}

	// iterate over all files, construct `file2layer`, `whreaddir`, `wh`
	for i, offset := range layers {
		if _, err := r.rd.Seek(offset, io.SeekStart); err != nil {
			return n, err
		}
		tr = tar.NewReader(r.rd)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return n, err
			}
			if hdr.Typeflag == tar.TypeDir {
				continue
			}

			// according to aufs documentation, whiteout files should be
			// hardlinks. I saw at least one docker container using regular
			// files for whiteouts.
			if hdr.Typeflag == tar.TypeLink || hdr.Typeflag == tar.TypeReg {
				basename := filepath.Base(hdr.Name)
				basedir := filepath.Dir(hdr.Name)
				if basename == _whReaddir {
					whreaddir[basedir] = i
					continue
				} else if strings.HasPrefix(basename, _whPrefix) {
					fname := strings.TrimPrefix(basename, _whPrefix)
					wh[filepath.Join(basedir, fname)] = i
					continue
				}
			}

			file2layer[hdr.Name] = i
		}
	}

	// construct directories to whiteout, for each layer.
	whIgnore := whiteoutDirs(whreaddir, len(layers))

	// iterate through all layers, all files, and write files.
	for i, offset := range layers {
		if _, err := r.rd.Seek(offset, io.SeekStart); err != nil {
			return n, err
		}
		tr = tar.NewReader(r.rd)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return n, err
			}
			if layer, ok := wh[hdr.Name]; ok && layer >= i {
				continue
			}
			if whIgnore[i].HasPrefix(hdr.Name) {
				continue
			}
			if hdr.Typeflag != tar.TypeDir && file2layer[hdr.Name] != i {
				continue
			}
			if err := writeFile(tr, tw, hdr); err != nil {
				return n, err
			}
		}
	}
	return n, nil
}

func writeFile(tr *tar.Reader, tw *tar.Writer, hdr *tar.Header) error {
	hdrOut := &tar.Header{
		Typeflag: hdr.Typeflag,
		Name:     hdr.Name,
		Linkname: hdr.Linkname,
		Size:     hdr.Size,
		Mode:     int64(hdr.Mode & 0777),
		Uid:      hdr.Uid,
		Gid:      hdr.Gid,
		Uname:    hdr.Uname,
		Gname:    hdr.Gname,
		ModTime:  hdr.ModTime,
		Devmajor: hdr.Devmajor,
		Devminor: hdr.Devminor,
		Format:   tar.FormatGNU,
	}

	if err := tw.WriteHeader(hdrOut); err != nil {
		return err
	}

	if hdr.Typeflag == tar.TypeReg {
		if _, err := io.Copy(tw, tr); err != nil {
			return err
		}
	}

	return nil
}

func whiteoutDirs(whreaddir map[string]int, nlayers int) []*tree {
	ret := make([]*tree, nlayers)
	for i := range ret {
		ret[i] = newTree()
	}
	for fname, layer := range whreaddir {
		if layer == 0 {
			continue
		}
		ret[layer-1].Add(fname)
	}
	for i := nlayers - 1; i > 0; i-- {
		ret[i-1].Merge(ret[i])
	}
	return ret
}

// byteCounter is an io.Writer that counts bytes written to it
type byteCounter struct {
	rw io.Writer
	n  int64
}

// Write writes to the underlying io.Writer and counts total written bytes
func (b *byteCounter) Write(data []byte) (n int, err error) {
	defer func() { b.n += int64(n) }()
	return b.rw.Write(data)
}
