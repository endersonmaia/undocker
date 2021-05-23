package main

import (
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/motiejus/code/undocker/internal/cmdlxcconfig"
	"github.com/motiejus/code/undocker/internal/cmdmanpage"
	"github.com/motiejus/code/undocker/internal/cmdrootfs"
)

func main() {
	parser := goflags.NewParser(nil, goflags.Default)

	rootfs := cmdrootfs.NewCommand()
	lxcconfig := cmdlxcconfig.NewCommand()
	manpage := cmdmanpage.NewCommand(parser)
	parser.AddCommand("rootfs", rootfs.ShortDesc(), rootfs.LongDesc(), rootfs)
	parser.AddCommand("lxcconfig", lxcconfig.ShortDesc(), lxcconfig.LongDesc(), lxcconfig)
	m, _ := parser.AddCommand("man-page", "", "", manpage)
	m.Hidden = true

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
