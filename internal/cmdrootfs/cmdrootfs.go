package cmdrootfs

import (
	"errors"
	"io"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/rootfs"
	"github.com/ulikunitz/xz"
	"go.uber.org/multierr"
)

// Command is "rootfs" command
type Command struct {
	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
	} `positional-args:"yes" required:"yes"`

	Xz bool `short:"J" long:"xz" description:"create XZ archive"`

	rootfsNew func(io.ReadSeeker) io.WriterTo
}

// Execute executes rootfs Command
func (c *Command) Execute(args []string) (err error) {
	if len(args) != 0 {
		return errors.New("too many args")
	}
	if c.rootfsNew == nil {
		c.rootfsNew = func(r io.ReadSeeker) io.WriterTo {
			return rootfs.New(r)
		}
	}

	rd, err := os.Open(string(c.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer func() { err = multierr.Append(err, rd.Close()) }()

	var out io.WriteCloser
	outf := string(c.PositionalArgs.Outfile)
	if outf == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(outf)
		if err != nil {
			return err
		}
	}
	defer func() { err = multierr.Append(err, out.Close()) }()

	if c.Xz {
		if out, err = xz.NewWriter(out); err != nil {
			return err
		}
		defer func() { err = multierr.Append(err, out.Close()) }()
	}

	if _, err := c.rootfsNew(rd).WriteTo(out); err != nil {
		return err
	}
	return nil
}
