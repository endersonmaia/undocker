package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~motiejus/code/undocker/rootfs"
	goflags "github.com/jessevdk/go-flags"
)

const _description = "Flatten a docker container image to a tarball"

func main() {
	parser := goflags.NewParser(newCommand(), goflags.Default)
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// command implements go-flags.Command
type command struct {
	flattener func(io.ReadSeeker, io.Writer) error
	Stdout    io.Writer

	PositionalArgs struct {
		Infile  goflags.Filename `long:"infile" description:"Input tarball"`
		Outfile string           `long:"outfile" description:"Output path, stdout is '-'"`
	} `positional-args:"yes" required:"yes"`
}

// newCommand creates a new Command struct
func newCommand() *command {
	return &command{
		flattener: rootfs.Flatten,
		Stdout:    os.Stdout,
	}
}

// Execute executes rootfs Command
func (c *command) Execute(args []string) (err error) {
	if len(args) != 0 {
		return errors.New("too many args")
	}

	rd, err := os.Open(string(c.PositionalArgs.Infile))
	if err != nil {
		return err
	}
	defer func() {
		err1 := rd.Close()
		if err == nil {
			err = err1
		}
	}()

	var out io.Writer
	if fname := string(c.PositionalArgs.Outfile); fname == "-" {
		out = c.Stdout
	} else {
		outf, err := os.Create(fname)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer func() {
			err1 := outf.Close()
			if err == nil {
				err = err1
			}
		}()
		out = outf
	}

	return c.flattener(rd, out)
}
