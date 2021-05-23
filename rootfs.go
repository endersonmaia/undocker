package main

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	goflags "github.com/jessevdk/go-flags"
)

type cmdRootFS struct {
	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output tarball (flattened file system)"`
	} `positional-args:"yes" required:"yes"`
}

const (
	_manifestJSON = "manifest.json"
)

func (r *cmdRootFS) Execute(args []string) error {
	if len(args) != 0 {
		return errors.New("too many args")
	}

	in, err := os.Open(string(r.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(string(r.PositionalArgs.Outfile))
	if err != nil {
		return err
	}
	defer out.Close()
	return r.rootfs(in, out)
}

type dockerManifestJSON []struct {
	Config string   `json:"Config"`
	Layers []string `json:"Layers"`
}

// rootfs accepts a docker layer tarball and writes it to outfile.
// 1. create map[string]io.ReadSeeker for each layer.
// 2. parse manifest.json and get the layer order.
// 3. go through each layer in order and write:
//    a) to an ordered slice: the file name.
//    b) to an FS map: where does the file come from?
//       I) layer name
//       II) offset (0 being the first file in the layer)
// 4. go through
func (r *cmdRootFS) rootfs(in io.ReadSeeker, out io.Writer) error {
	tr := tar.NewReader(in)
	layerOffsets := map[string]int64{}
	var manifest dockerManifestJSON

	// phase 1: get layer offsets and manifest.json
	for {
		hdrIn, err := tr.Next()
		if err == io.EOF {
			break
		}

		if hdrIn.Typeflag != tar.TypeReg {
			continue
		}

		switch {
		case hdrIn.Name == _manifestJSON:
			dec := json.NewDecoder(tr)
			if err := dec.Decode(&manifest); err != nil {
				return fmt.Errorf("decode manifest.json: %w", err)
			}
		case strings.HasSuffix(hdrIn.Name, "/layer.tar"):
			here, err := in.Seek(0, io.SeekCurrent)
			if err != nil {
				return fmt.Errorf("seek: %w", err)
			}
			layers[hdrIn.Name] = here
		}
	}

	// phase 1.5: enumerate layers
	layers := make([]int64, len(layers))
	for i, name := range manifest.Layers {
		layers[i] = layerOffsets[name]
	}

	// phase 2: iterate through all layers and extract filenames
	// to files and ordered
	files := map[string]struct{
		layer string

	fmt.Printf("layers: %+v\n", layers)

	return nil
	//return tw.Close()
}
