package cmd

import (
	"io"
	"os"
)

// BaseCommand provides common fields to all commands.
type BaseCommand struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Init initializes BaseCommand with default arguments
func (b *BaseCommand) Init() {
	b.Stdin = os.Stdin
	b.Stdout = os.Stdout
	b.Stderr = os.Stderr
}
