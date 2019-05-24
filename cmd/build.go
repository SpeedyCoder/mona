package cmd

import (
	"github.com/davidsbond/mona/internal/command"

	"github.com/urfave/cli"
)

// Build generates a cli command that builds any modified modules within
// the project.
func Build() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "Builds any modified modules within the project",
		Action: func(ctx *cli.Context) error {
			return command.Build()
		},
	}
}