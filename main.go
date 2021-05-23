package main

import (
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmdlxcconfig"
	"github.com/motiejus/code/undocker/internal/cmdrootfs"
)

type (
	params struct {
		RootFS    cmdrootfs.Command    `command:"rootfs" description:"Unpack a docker container image to a single filesystem tarball"`
		LXCConfig cmdlxcconfig.Command `command:"lxcconfig" description:"Create an LXC-compatible container configuration"`
	}
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
