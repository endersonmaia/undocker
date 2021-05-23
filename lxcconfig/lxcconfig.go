package lxcconfig

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
)

const (
	_json         = ".json"
	_manifestJSON = "manifest.json"
)

var (
	_lxcTemplate = template.Must(
		template.New("lxcconfig").Parse("" +
			"lxc.include = LXC_TEMPLATE_CONFIG/common.conf\n" +
			"lxc.architecture = {{ .Architecture }}\n" +
			"lxc.execute.cmd = '{{ .Cmd }}'\n" +
			"{{ if .Cwd }}lxc.init.cwd = {{ .Cwd }}\n{{ end }}" +
			"{{ range .Env }}lxc.environment = {{ . }}\n{{ end }}"))
	errBadManifest = errors.New("bad or missing manifest.json")
)

type (
	// lxcConfig is passed to _lxcTemplate
	lxcConfig struct {
		Architecture string
		Cmd          string
		Cwd          string
		Env          []string
	}

	// offsetSize is a tuple which stores file offset and size
	offsetSize struct {
		offset int64
		size   int64
	}

	// dockerManifest is manifest.json
	dockerManifest []struct {
		Config string `json:"Config"`
	}

	// dockerConfig returns interesting configs for the container. user/group are
	// skipped, since Docker allows specifying them by name, which would require
	// peeking into the container image's /etc/passwd to resolve the names to ints.
	dockerConfig struct {
		Architecture string             `json:"architecture"`
		Config       dockerConfigConfig `json:"config"`
	}

	dockerConfigConfig struct {
		Entrypoint []string `json:"Entrypoint"`
		Cmd        []string `json:"Cmd"`
		WorkingDir string   `json:"WorkingDir"`
		Env        []string `json:"Env"`
	}
)

// LXCConfig accepts a Docker container image and returns lxc configuration.
func LXCConfig(rd io.ReadSeeker, wr io.Writer) error {
	dockerCfg, err := getDockerConfig(rd)
	if err != nil {
		return err
	}
	lxcCfg := docker2lxc(dockerCfg)
	return lxcCfg.WriteTo(wr)
}

func docker2lxc(d dockerConfig) lxcConfig {
	// cmd/entrypoint logic is copied from lxc-oci template
	ep := strings.Join(d.Config.Entrypoint, " ")
	cmd := strings.Join(d.Config.Cmd, " ")
	if len(ep) == 0 {
		ep = cmd
		if len(ep) == 0 {
			ep = "/bin/sh"
		}
	} else if len(cmd) != 0 {
		ep = ep + " " + cmd
	}
	return lxcConfig{
		Architecture: d.Architecture,
		Cmd:          ep,
		Env:          d.Config.Env,
		Cwd:          d.Config.WorkingDir,
	}
}

func (l lxcConfig) WriteTo(wr io.Writer) error {
	return _lxcTemplate.Execute(wr, l)
}

func getDockerConfig(rd io.ReadSeeker) (dockerConfig, error) {
	tr := tar.NewReader(rd)
	// get offsets to all json files rd the archive
	jsonOffsets := map[string]offsetSize{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(hdr.Name, _json) {
			continue
		}
		here, err := rd.Seek(0, io.SeekCurrent)
		if err != nil {
			return dockerConfig{}, err
		}
		jsonOffsets[hdr.Name] = offsetSize{
			offset: here,
			size:   hdr.Size,
		}
	}

	// manifest is the docker manifest rd the image
	var manifest dockerManifest
	if err := parseJSON(rd, jsonOffsets, _manifestJSON, &manifest); err != nil {
		return dockerConfig{}, err
	}
	if len(manifest) == 0 {
		return dockerConfig{}, errBadManifest
	}

	var config dockerConfig
	if err := parseJSON(rd, jsonOffsets, manifest[0].Config, &config); err != nil {
		return dockerConfig{}, err
	}

	return config, nil
}

func parseJSON(rd io.ReadSeeker, offsets map[string]offsetSize, fname string, c interface{}) error {
	osize, ok := offsets[fname]
	if !ok {
		return fmt.Errorf("file %s not found", fname)
	}
	if _, err := rd.Seek(osize.offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek to %s: %w", fname, err)
	}

	lrd := io.LimitReader(rd, osize.size)
	dec := json.NewDecoder(lrd)
	if err := dec.Decode(c); err != nil {
		return fmt.Errorf("decode %s: %w", fname, err)
	}
	return nil
}
