package main

import (
	"errors"

	goflags "github.com/jessevdk/go-flags"
)

type cmdRootFS struct {
	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output tarball (flattened file system)"`
	} `positional-args:"yes" required:"yes"`
}

func (r *cmdRootFS) Execute(args []string) error {
	if len(args) != 0 {
		return errors.New("too many args")
	}
	return nil
}
