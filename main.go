// Package main is a simple command-line application on top of rootfs.Flatten.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"git.sr.ht/~motiejus/undocker/rootfs"
)

const _usage = `Usage:
  %s <infile> <outfile>

Flatten a Docker container image to a root file system.

Arguments:
  <infile>:  Input Docker container. Tarball.
  <outfile>: Output tarball, the root file system. '-' is stdout.
`

func main() {
	runtime.GOMAXPROCS(1) // no need to create that many threads

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, _usage, filepath.Base(os.Args[0]))
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

func (c *command) execute(infile string, outfile string) (err error) {
	rd, err := os.Open(infile)
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
	if outfile == "-" {
		out = c.Stdout
	} else {
		outf, err := os.Create(outfile)
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
