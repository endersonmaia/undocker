package cmdlxcconfig

import (
	"errors"
	"fmt"
	"io"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/lxcconfig"
	"go.uber.org/multierr"
)

// Command is "lxcconfig" command
type (
	Command struct {
		Stdout io.Writer

		PositionalArgs struct {
			Infile  goflags.Filename `long:"infile" description:"Input tarball"`
			Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
		} `positional-args:"yes" required:"yes"`
	}
)

func NewCommand() *Command {
	return &Command{
		Stdout: os.Stdout,
	}
}

func (*Command) ShortDesc() string { return "Create an LXC-compatible container configuration" }
func (*Command) LongDesc() string  { return "" }

// Execute executes lxcconfig Command
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
	outf := string(c.PositionalArgs.Outfile)
	if fname := string(c.PositionalArgs.Outfile); fname == "-" {
		out = c.Stdout
	} else {
		outf, err := os.Create(outf)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer func() { err = multierr.Append(err, outf.Close()) }()
		out = outf
	}

	return lxcconfig.LXCConfig(rd, out)
}
