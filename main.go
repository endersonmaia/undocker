package main

import (
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmdlxcconfig"
	"github.com/motiejus/code/undocker/internal/cmdrootfs"
)

func main() {
	flags := goflags.NewParser(nil, goflags.Default)

	rootfs := cmdrootfs.NewCommand()
	lxcconfig := cmdlxcconfig.NewCommand()
	flags.AddCommand("rootfs", rootfs.ShortDesc(), rootfs.LongDesc(), rootfs)
	flags.AddCommand("lxcconfig", lxcconfig.ShortDesc(), lxcconfig.LongDesc(), lxcconfig)

	_, err := flags.Parse()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
