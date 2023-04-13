// Package main is a simple command-line application on top of rootfs.Flatten.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"git.jakstys.lt/motiejus/undocker/rootfs"
)

var Version = "unknown"
var VersionHash = "unknown"

const _usage = `Usage:
  %s <infile> <outfile>

Flatten a Docker container image to a root file system.

Arguments:
  <infile>:  Input Docker container. Tarball.
  <outfile>: Output tarball, the root file system. '-' is stdout.

undocker %s (%s)
Built with %s
`

func main() {
	runtime.GOMAXPROCS(1) // no need to create that many threads

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, _usage,
			filepath.Base(os.Args[0]),
			Version,
			VersionHash,
			runtime.Version(),
		)
		os.Exit(1)
	}

	c := &command{flattener: rootfs.Flatten, Stdout: os.Stdout}
	if err := c.execute(os.Args[1], os.Args[2]); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type command struct {
	flattener func(io.ReadSeeker, io.Writer) error
	Stdout    io.Writer
}

func (c *command) execute(infile string, outfile string) (_err error) {
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

	return c.flattener(rd, out)
}
