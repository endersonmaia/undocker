package cmdrootfs

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	var _foo = []byte("foo foo")

	tests := []struct {
		name    string
		fixture func(*testing.T, string)
		infile  string
		outfile string
		wantErr string
	}{
		{
			name:   "ok passthrough via stdout",
			infile: "t10-in.txt",
			fixture: func(t *testing.T, dir string) {
				fname := filepath.Join(dir, "t10-in.txt")
				require.NoError(t, ioutil.WriteFile(fname, _foo, 0644))
			},
			outfile: "-",
		},
		{
			name:   "ok passthrough via file",
			infile: "t20-in.txt",
			fixture: func(t *testing.T, dir string) {
				fname := filepath.Join(dir, "t20-in.txt")
				require.NoError(t, ioutil.WriteFile(fname, _foo, 0644))
			},
			outfile: "t20-out.txt",
		},
		{
			name:    "infile does not exist",
			infile:  "t3-does-not-exist.txt",
			wantErr: "^open .*t3-does-not-exist.txt: no such file or directory$",
		},
		{
			name:    "outpath dir not writable",
			outfile: filepath.Join("t4", "does", "not", "exist"),
			wantErr: "^create: open .*/t4/does/not/exist: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			var stdout bytes.Buffer
			c := &Command{BaseCommand: cmd.BaseCommand{Stdout: &stdout}}
			if tt.fixture != nil {
				tt.fixture(t, dir)
			}
			if tt.outfile != "-" {
				tt.outfile = filepath.Join(dir, tt.outfile)
			}
			inf := filepath.Join(dir, tt.infile)
			c.PositionalArgs.Infile = goflags.Filename(inf)
			c.PositionalArgs.Outfile = tt.outfile
			c.rootfsNew = func(r io.ReadSeeker) io.WriterTo {
				return &passthrough{r}
			}

			err := c.Execute(nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Regexp(t, tt.wantErr, err.Error())
				return
			}
			var out []byte
			require.NoError(t, err)
			if tt.outfile == "-" {
				out = stdout.Bytes()
			} else {
				out, err = ioutil.ReadFile(tt.outfile)
				require.NoError(t, err)
			}
			assert.Equal(t, []byte("foo foo"), out)
		})
	}
}

type passthrough struct{ r io.Reader }

func (p *passthrough) WriteTo(w io.Writer) (int64, error) { return io.Copy(w, p.r) }
