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
	Config string   `json:"Config,omitempty"`
	Layers []string `json:"Layers"`
}

// RootFS accepts a docker layer tarball and writes it to outfile.
// 1. create map[string]io.ReadSeeker for each layer.
// 2. parse manifest.json and get the layer order.
// 3. go through each layer in order and write:
//    a) to an ordered slice: the file name.
//    b) to an FS map: where does the file come from?
//       I) layer name
//       II) offset (0 being the first file in the layer)
// 4. go through
func RootFS(in io.ReadSeeker, out io.Writer) (err error) {
	tr := tar.NewReader(in)
	tw := tar.NewWriter(out)
	defer func() { err = multierr.Append(err, tw.Close()) }()

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
				return fmt.Errorf("parse %s: %w", _manifestJSON, err)
			}
		case strings.HasSuffix(hdr.Name, _layerSuffix):
			here, err := in.Seek(0, io.SeekCurrent)
			if err != nil {
				return err
			}
			layerOffsets[hdr.Name] = here
		}
	}

	if len(manifest) == 0 || len(layerOffsets) != len(manifest[0].Layers) {
		return errBadManifest
	}

	// enumerate layers the way they would be laid down in the image
	layers := make([]int64, len(layerOffsets))
	for i, name := range manifest[0].Layers {
		layers[i] = layerOffsets[name]
	}

	// file2layer maps a filename to layer number (index in "layers")
	file2layer := map[string]int{}

	// whreaddir maps a directory to a layer number until which
	// its contents should be ignored, exclusively.
	whreaddir := map[string]int{}

	// wh maps a filename to a layer until which it should be ignored,
	// inclusively.
	wh := map[string]int{}

	// build up `file2layer`, `whreaddir`, `wh`
	for i, offset := range layers {
		if _, err := in.Seek(offset, io.SeekStart); err != nil {
			return err
		}
		tr = tar.NewReader(in)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if hdr.Typeflag == tar.TypeDir {
				continue
			}

			if hdr.Typeflag == tar.TypeLink {
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

	// phase 3: iterate through all layers and write files.
	for i, offset := range layers {
		if _, err := in.Seek(offset, io.SeekStart); err != nil {
			return err
		}
		tr = tar.NewReader(in)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if file2layer[hdr.Name] != i {
				continue
			}
			if err := writeFile(tr, tw, hdr); err != nil {
				return err
			}
		}
	}
	return nil
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
