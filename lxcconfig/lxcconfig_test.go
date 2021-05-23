package lxcconfig

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/motiejus/code/undocker/internal/tartest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLXCConfig(t *testing.T) {
	tests := []struct {
		name   string
		docker dockerConfig
		want   string
	}{
		{
			name: "just architecture",
			docker: dockerConfig{
				Architecture: "amd64",
			},
			want: strings.Join([]string{
				`lxc.include = LXC_TEMPLATE_CONFIG/common.conf`,
				`lxc.architecture = amd64`,
				`lxc.init.cmd = '/bin/sh'`,
				``,
			}, "\n"),
		},
		{
			name: "all fields",
			docker: dockerConfig{
				Architecture: "amd64",
				Config: dockerConfigConfig{
					Entrypoint: []string{"/entrypoint.sh"},
					Cmd:        []string{"/bin/sh", "-c", "echo foo"},
					WorkingDir: "/x",
					Env: []string{
						`LONGNAME="Foo Bar"`,
						"SHELL=/bin/tcsh",
					},
				},
			},
			want: strings.Join([]string{
				`lxc.include = LXC_TEMPLATE_CONFIG/common.conf`,
				`lxc.architecture = amd64`,
				`lxc.init.cmd = '/entrypoint.sh /bin/sh -c "echo foo"'`,
				`lxc.init.cwd = /x`,
				`lxc.environment = LONGNAME="Foo Bar"`,
				`lxc.environment = SHELL=/bin/tcsh`,
				``,
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := tartest.File{
				Name:     "manifest.json",
				Contents: bytes.NewBufferString(`[{"Config":"config.json"}]`),
			}
			archive := tartest.Tarball{
				manifest,
				tt.docker,
			}
			in := bytes.NewReader(archive.Buffer().Bytes())
			var buf bytes.Buffer
			require.NoError(t, LXCConfig(in, &buf))
			assert.Equal(t, tt.want, string(buf.Bytes()))
		})
	}
}

// Helpers
func (c dockerConfig) Tar(tw *tar.Writer) error {
	configJSON, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return tartest.File{
		Name:     "config.json",
		Contents: bytes.NewBuffer(configJSON),
	}.Tar(tw)
}
