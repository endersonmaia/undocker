package cmdmanpage

import (
	"io"
	"os"

	goflags "github.com/jessevdk/go-flags"
)

type Command struct {
	parser *goflags.Parser
	stdout io.Writer
}

func NewCommand(parser *goflags.Parser) *Command {
	return &Command{
		parser: parser,
		stdout: os.Stdout,
	}
}

func (c *Command) Execute(args []string) error {
	c.parser.WriteManPage(c.stdout)
	return nil
}
