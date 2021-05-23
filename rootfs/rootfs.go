package rootfs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/motiejus/code/undocker/internal/bytecounter"
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

type (
	// RootFS accepts a docker layer tarball and flattens it.
	RootFS struct {
		rd io.ReadSeeker
	}

	dockerManifestJSON []struct {
		Layers []string `json:"Layers"`
	}

	nameOffset struct {
		name   string
		offset int64
	}
)

// New creates a new RootFS'er.
func New(rd io.ReadSeeker) *RootFS {
	return &RootFS{rd: rd}
}

// WriteTo writes a docker image to an open tarball.
func (r *RootFS) WriteTo(w io.Writer) (n int64, err error) {
	tr := tar.NewReader(r.rd)
	var closer func() error

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
	layers := make([]nameOffset, len(layerOffsets))
	for i, name := range manifest[0].Layers {
		layers[i] = nameOffset{
			name:   name,
			offset: layerOffsets[name],
		}
	}

	// file2layer maps a filename to layer number (index in "layers")
	file2layer := map[string]int{}

	// whreaddir maps `wh..wh..opq` file to a layer; see doc.go
	whreaddir := map[string]int{}

	// wh maps a filename to a layer until which it should be ignored,
	// inclusively; see doc.go
	wh := map[string]int{}

	// iterate over all files, construct `file2layer`, `whreaddir`, `wh`
	for i, no := range layers {
		if _, err := r.rd.Seek(no.offset, io.SeekStart); err != nil {
			return n, err
		}
		tr, closer = openTargz(r.rd)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return n, fmt.Errorf("decode %s: %w", no.name, err)
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
		if err := closer(); err != nil {
			return n, err
		}
	}

	// construct directories to whiteout, for each layer.
	whIgnore := whiteoutDirs(whreaddir, len(layers))

	wr := bytecounter.New(w)
	tw := tar.NewWriter(wr)
	defer func() {
		err = multierr.Append(err, tw.Close())
		n = wr.N
	}()
	// iterate through all layers, all files, and write files.
	for i, no := range layers {
		if _, err := r.rd.Seek(no.offset, io.SeekStart); err != nil {
			return n, err
		}
		tr, closer = openTargz(r.rd)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return n, fmt.Errorf("decode %s: %w", no.name, err)
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
		if err := closer(); err != nil {
			return n, err
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

// openTargz creates a tar reader from a targzip or tar.
//
// We may be looking at magic values for tar and/or gzip,
// which would mean "cleaner" code (e.g. no proxyWriter),
// but that would mean re-implementing gzip.readHeader(),
// which is ... already in stdlib.
func openTargz(r io.Reader) (*tar.Reader, func() error) {
	hdrbuf := &bytes.Buffer{}
	hdrw := &proxyWriter{w: hdrbuf}
	gz, err := gzip.NewReader(io.TeeReader(r, hdrw))
	if err == nil {
		hdrw.w = ioutil.Discard
		hdrbuf = nil
		return tar.NewReader(gz), gz.Close
	}
	return tar.NewReader(io.MultiReader(hdrbuf, r)), func() error { return nil }
}

// proxyWriter is a pass-through writer. Its underlying writer can be changed
// on-the-fly. Useful when there is a stream that needs to be discarded (change
// the underlying writer to, say, ioutil.Discard).
type proxyWriter struct {
	w io.Writer
}

// Write writes a slice to the underlying w.
func (pw *proxyWriter) Write(p []byte) (int, error) {
	return pw.w.Write(p)
}
