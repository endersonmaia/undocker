package main

import (
	"errors"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/rootfs"
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
		Outfile string           `long:"outfile" description:"Output tarball (flattened file system)"`
	} `positional-args:"yes" required:"yes"`
}

func (r *cmdRootFS) Execute(args []string) error {
	if len(args) != 0 {
		return errors.New("too many args")
	}

	in, err := os.Open(string(r.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(string(r.PositionalArgs.Outfile))
	if err != nil {
		return err
	}
	defer out.Close()
	return rootfs.Rootfs(in, out)
}
