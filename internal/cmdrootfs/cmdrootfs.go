package cmdrootfs

import (
	"errors"
	"fmt"
	"io"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmd"
	"github.com/motiejus/code/undocker/rootfs"
	"go.uber.org/multierr"
)

type (

	// Command is "rootfs" command
	Command struct {
		cmd.BaseCommand

		PositionalArgs struct {
			Infile  goflags.Filename `long:"infile" description:"Input tarball"`
			Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
		} `positional-args:"yes" required:"yes"`

		flattener func(io.ReadSeeker, io.Writer) error
	}
)

// Execute executes rootfs Command
func (c *Command) Execute(args []string) (err error) {
	if len(args) != 0 {
		return errors.New("too many args")
	}
	if c.flattener == nil {
		c.init()
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

// init() initializes Command with the default options.
//
// Since constructors for sub-commands requires lots of boilerplate,
// command will initialize itself.
func (c *Command) init() {
	c.BaseCommand.Init()
	c.flattener = rootfs.Flatten
}
