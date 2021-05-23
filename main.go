package main

import (
	"fmt"
	"os"
	"errors"

	goflags "github.com/jessevdk/go-flags"
)

type opts struct {
	PositionalArgs struct {
		Infile     goflags.Filename `long:"infile" description:"Docker container tarball"`
		Outfile string           `long:"outfile" description:"Output tarball"`
	} `positional-args:"yes" required:"yes"`
}

func main() {
	if err := run(os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func run(args []string) error {
	var flags opts
	args1, err := goflags.ParseArgs(&flags, args)
	if err != nil {
		return err
	}
	if len(args1) != 0 {
		return errors.New("too many args")
	}

	fmt.Printf("infile: %s\n", flags.PositionalArgs.Infile)
	fmt.Printf("outfile: %s\n", flags.PositionalArgs.Outfile)

	return nil
}
