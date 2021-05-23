package cmdrootfs

import (
	"errors"
	"io"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmd"
	"github.com/motiejus/code/undocker/rootfs"
	"github.com/ulikunitz/xz"
	"go.uber.org/multierr"
)

// Command is "rootfs" command
type Command struct {
	cmd.BaseCommand

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
		c.init()
	}

	rd, err := os.Open(string(c.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer func() { err = multierr.Append(err, rd.Close()) }()

	var out io.Writer
	var outf *os.File
	if fname := string(c.PositionalArgs.Outfile); fname == "-" {
		outf = os.Stdout
	} else {
		outf, err = os.Create(fname)
		if err != nil {
			return err
		}
	}
	out = outf
	defer func() { err = multierr.Append(err, outf.Close()) }()

	if c.Xz {
		outz, err := xz.NewWriter(out)
		if err != nil {
			return err
		}
		defer func() { err = multierr.Append(err, outz.Close()) }()
		out = outz
	}

	if _, err := c.rootfsNew(rd).WriteTo(out); err != nil {
		return err
	}
	return nil
}

// init() initializes Command with the default options.
//
// Since constructors for sub-commands requires lots of boilerplate,
// command will initialize itself.
func (c *Command) init() {
	c.BaseCommand.Init()
	c.rootfsNew = func(r io.ReadSeeker) io.WriterTo {
		return rootfs.New(r)
	}
}
