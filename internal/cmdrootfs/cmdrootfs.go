package cmdrootfs

import (
	"errors"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/rootfs"
	"go.uber.org/multierr"
)

type CmdRootFS struct {
	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
	} `positional-args:"yes" required:"yes"`
}

func (r *CmdRootFS) Execute(args []string) (err error) {
	if len(args) != 0 {
		return errors.New("too many args")
	}

	rd, err := os.Open(string(r.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer func() { err = multierr.Append(err, rd.Close()) }()

	var out *os.File
	outf := string(r.PositionalArgs.Outfile)
	if outf == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(outf)
		if err != nil {
			return err
		}
	}
	defer func() { err = multierr.Append(err, out.Close()) }()

	return rootfs.New(rd).WriteTo(out)
}
