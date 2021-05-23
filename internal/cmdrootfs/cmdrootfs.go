package cmdrootfs

import (
	"errors"
	"fmt"
	"io"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/rootfs"
	"go.uber.org/multierr"
)

const _description = "Flatten a docker container image to a tarball"

type (
	// Command is an implementation of go-flags.Command
	Command struct {
		flattener func(io.ReadSeeker, io.Writer) error
		Stdout    io.Writer

		PositionalArgs struct {
			Infile  goflags.Filename `long:"infile" desc:"Input tarball"`
			Outfile string           `long:"outfile" desc:"Output path, stdout is '-'"`
		} `positional-args:"yes" required:"yes"`
	}
)

func NewCommand() *Command {
	return &Command{
		flattener: rootfs.Flatten,
		Stdout:    os.Stdout,
	}
}

func (*Command) ShortDesc() string { return _description }
func (*Command) LongDesc() string  { return _description }

// Execute executes rootfs Command
func (c *Command) Execute(args []string) (err error) {
	if len(args) != 0 {
		return errors.New("too many args")
	}

	rd, err := os.Open(string(c.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer func() { err = multierr.Append(err, rd.Close()) }()

	var out io.Writer
	if fname := string(c.PositionalArgs.Outfile); fname == "-" {
		out = c.Stdout
	} else {
		outf, err := os.Create(fname)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer func() { err = multierr.Append(err, outf.Close()) }()
		out = outf
	}

	return c.flattener(rd, out)
}
