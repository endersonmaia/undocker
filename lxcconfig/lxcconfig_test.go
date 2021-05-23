package lxcconfig

import (
	"bytes"
	"testing"

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
			want: `lxc.include = LXC_TEMPLATE_CONFIG/common.conf
lxc.architecture = amd64
lxc.execute.cmd = '/bin/sh'
`,
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
			want: `lxc.include = LXC_TEMPLATE_CONFIG/common.conf
lxc.architecture = amd64
lxc.execute.cmd = '/entrypoint.sh /bin/sh -c echo foo'
lxc.init.cwd = /x
lxc.environment = LONGNAME="Foo Bar"
lxc.environment = SHELL=/bin/tcsh
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, docker2lxc(tt.docker).WriteTo(&buf))
			assert.Equal(t, tt.want, string(buf.Bytes()))
		})
	}
}
