// Package main is a simple command-line application on top of rootfs.Flatten.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"git.sr.ht/~motiejus/undocker/rootfs"
)

var Version = "unknown"
var VersionHash = "unknown"

const _usage = `Usage:
  %s [OPTION]... <infile> <outfile>

Flatten a Docker container image to a root file system.

Arguments:
  <infile>   Input Docker container. Tarball.
  <outfile>  Output tarball, the root file system. '-' is stdout.

Options:
  --prefix=<prefix>  prefix all destination files with a given string.

undocker %s (%s)
Built with %s
`

func usage(pre string, out io.Writer) {
	fmt.Fprintf(out, pre+_usage,
		filepath.Base(os.Args[0]),
		Version,
		VersionHash,
		runtime.Version(),
	)
}

func usageErr(pre string) {
	usage(pre, os.Stderr)
	os.Exit(2)
}

func main() {
	runtime.GOMAXPROCS(1) // no need to create that many threads

	var filePrefix string
	fs := flag.NewFlagSet("undocker", flag.ExitOnError)
	fs.Usage = func() { usageErr("") }
	fs.StringVar(&filePrefix, "prefix", "", "prefix files in the tarball")

	if len(os.Args) == 1 {
		usageErr("")
	}

	_ = fs.Parse(os.Args[1:]) // ExitOnError captures it

	args := fs.Args()
	if len(args) != 2 {
		usageErr("invalid number of arguments\n")
	}

	c := &command{flattener: rootfs.Flatten, Stdout: os.Stdout}
	if err := c.execute(args[0], args[1], filePrefix); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type command struct {
	flattener func(io.ReadSeeker, io.Writer, ...rootfs.Option) error
	Stdout    io.Writer
}

func (c *command) execute(infile, outfile, filePrefix string) (_err error) {
	rd, err := os.Open(infile)
	if err != nil {
		return err
	}
	defer func() {
		err := rd.Close()
		if _err == nil {
			_err = err
		}
	}()

	var out io.Writer
	if outfile == "-" {
		out = c.Stdout
	} else {
		outf, err := os.Create(outfile)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer func() {
			err := outf.Close()
			if _err != nil {
				os.Remove(outfile)
			} else {
				_err = err
			}
		}()
		out = outf
	}

	return c.flattener(rd, out, rootfs.WithFilePrefix(filePrefix))
}
