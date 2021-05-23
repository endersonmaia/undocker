package rootfs

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	_manifestJSON = "manifest.json"
)

type dockerManifestJSON []struct {
	Config string   `json:"Config"`
	Layers []string `json:"Layers"`
}

// Rootfs accepts a docker layer tarball and writes it to outfile.
// 1. create map[string]io.ReadSeeker for each layer.
// 2. parse manifest.json and get the layer order.
// 3. go through each layer in order and write:
//    a) to an ordered slice: the file name.
//    b) to an FS map: where does the file come from?
//       I) layer name
//       II) offset (0 being the first file in the layer)
// 4. go through
func Rootfs(in io.ReadSeeker, out io.Writer) error {
	tr := tar.NewReader(in)
	tw := tar.NewWriter(out)
	// layerOffsets maps a layer name (a9b123c0daa/layer.tar) to it's offset
	layerOffsets := map[string]int64{}

	// manifest is the docker manifest in the image
	var manifest dockerManifestJSON

	// phase 1: get layer offsets and manifest.json
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		switch {
		case hdr.Name == _manifestJSON:
			dec := json.NewDecoder(tr)
			if err := dec.Decode(&manifest); err != nil {
				return fmt.Errorf("decode manifest.json: %w", err)
			}
		case strings.HasSuffix(hdr.Name, "/layer.tar"):
			here, err := in.Seek(0, io.SeekCurrent)
			if err != nil {
				return fmt.Errorf("seek: %w", err)
			}
			layerOffsets[hdr.Name] = here
		}
	}

	// phase 1.5: enumerate layers
	layers := make([]int64, len(layerOffsets))
	for i, name := range manifest[0].Layers {
		layers[i] = layerOffsets[name]
	}

	// file2layer maps a filename to layer number (index in "layers")
	file2layer := map[string]int{}

	// phase 2: iterate through all layers and save filenames
	// for all kinds of files.
	for i, offset := range layers {
		if _, err := in.Seek(offset, io.SeekStart); err != nil {
			fmt.Errorf("seek: %w", err)
		}
		tr = tar.NewReader(in)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			file2layer[hdr.Name] = i
		}
	}

	// phase 3: iterate through all layers and write files.
	for i, offset := range layers {
		if _, err := in.Seek(offset, io.SeekStart); err != nil {
			fmt.Errorf("seek: %w", err)
		}
		tr = tar.NewReader(in)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if file2layer[hdr.Name] != i {
				continue
			}

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
		}
	}

	return tw.Close()
}
