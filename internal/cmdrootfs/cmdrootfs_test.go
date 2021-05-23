package cmdrootfs

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		in      []byte
		infile  string
		outfile string
		want    []byte
		wantErr string
	}{
		{
			name:    "ok passthrough via stdout",
			in:      []byte("foo"),
			outfile: "-",
			want:    []byte("foo"),
		},
		{
			name:    "ok passthrough via file",
			in:      []byte("foo"),
			outfile: filepath.Join(dir, "t1.txt"),
			want:    []byte("foo"),
		},
		{
			name:    "infile does not exist",
			infile:  filepath.Join(dir, "does", "not", "exist"),
			wantErr: "open <...>/does/not/exist: enoent",
		},
		{
			name:    "outpath dir not writable",
			outfile: filepath.Join(dir, "does", "not", "exist"),
			wantErr: "create: stat <...>/does/not/exist: enoent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			c := &Command{BaseCommand: cmd.BaseCommand{Stdout: &stdout}}
			c.PositionalArgs.Infile = goflags.Filename(tt.infile)
			c.PositionalArgs.Outfile = tt.outfile
			c.rootfsNew = func(r io.ReadSeeker) io.WriterTo {
				return &passthrough{r}
			}

			err := c.Execute(nil)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, stdout.Bytes())
		})
	}
}

type passthrough struct{ r io.Reader }

func (p *passthrough) WriteTo(w io.Writer) (int64, error) { return io.Copy(w, p.r) }
