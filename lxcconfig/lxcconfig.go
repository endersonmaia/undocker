package lxcconfig

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const _manifestJSON = "manifest.json"

// dockerManifest is manifest.json
type dockerManifest []struct {
	Config string `json:"Config"`
}

// dockerConfig returns interesting configs for the container. user/group are
// skipped, since Docker allows specifying them by name, which would require
// peeking into the container image's /etc/passwd to resolve the names to ints.
type dockerConfig struct {
	Architecture string `json:"architecture"`
	Config       struct {
		Entrypoint []string `json:"Entrypoint"`
		Cmd        []string `json:"Cmd"`
		Env        []string `json:"Env"`
		WorkingDir string   `json:"WorkingDir"`
	} `json:"config"`
}

var (
	errBadManifest = errors.New("bad or missing manifest.json")
)

func getDockerConfig(in io.ReadSeeker) (dockerConfig, error) {
	tr := tar.NewReader(in)
	// get offsets to all json files in the archive
	jsonOffsets := map[string]int64{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(".json", hdr.Name) {
			continue
		}
		here, err := in.Seek(0, io.SeekCurrent)
		if err != nil {
			return dockerConfig{}, err
		}
		jsonOffsets[hdr.Name] = here
	}

	// manifest is the docker manifest in the image
	var manifest dockerManifest
	if err := parseJSON(in, jsonOffsets, _manifestJSON, &manifest); err != nil {
		return dockerConfig{}, err
	}
	if len(manifest) == 0 {
		return dockerConfig{}, errBadManifest
	}

	var config dockerConfig
	if err := parseJSON(in, jsonOffsets, manifest[0].Config, &config); err != nil {
		return dockerConfig{}, err
	}

	return config, nil
}

func parseJSON(in io.ReadSeeker, offsets map[string]int64, fname string, c interface{}) error {
	configOffset, ok := offsets[fname]
	if !ok {
		return fmt.Errorf("file %s not found", fname)
	}
	if _, err := in.Seek(configOffset, io.SeekStart); err != nil {
		return fmt.Errorf("seek to %s: %w", fname, err)
	}
	tr := tar.NewReader(in)
	if _, err := tr.Next(); err != nil {
		return err
	}
	dec := json.NewDecoder(tr)
	if err := dec.Decode(c); err != nil {
		return fmt.Errorf("decode %s: %w", fname, err)
	}
	return nil
}
