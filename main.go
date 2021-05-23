package main

import (
	"os"

	goflags "github.com/jessevdk/go-flags"
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
