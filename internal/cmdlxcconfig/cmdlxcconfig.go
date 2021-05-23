package cmdlxcconfig

import (
	"errors"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/lxcconfig"
	"go.uber.org/multierr"
)

// Command is "lxcconfig" command
type Command struct {
	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
	} `positional-args:"yes" required:"yes"`
}

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

	var out *os.File
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

	return lxcconfig.LXCConfig(rd, out)
}
