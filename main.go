package main

import (
	"errors"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/rootfs"
	"go.uber.org/multierr"
)

type (
	params struct {
		RootFS   cmdRootFS   `command:"rootfs" description:"Unpack a docker container image to a single filesystem tarball"`
		Manifest cmdManifest `command:"manifest"`
	}

	cmdManifest struct{} // stub
)

func main() {
	if err := run(os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func run(args []string) error {
	var opts params
	if _, err := goflags.ParseArgs(&opts, args[1:]); err != nil {
		return err
	}

	return nil
}

type cmdRootFS struct {
	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
	} `positional-args:"yes" required:"yes"`
}

func (r *cmdRootFS) Execute(args []string) (err error) {
	if len(args) != 0 {
		return errors.New("too many args")
	}

	in, err := os.Open(string(r.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer func() { err = multierr.Append(err, in.Close()) }()

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

	return rootfs.RootFS(in, out)
}
