package cmd

import (
	"github.com/davidsbond/mona/internal/command"
	"github.com/davidsbond/mona/internal/config"
	"github.com/davidsbond/mona/internal/deps"
	"github.com/urfave/cli"
)

// Build generates a cli command that builds any modified apps within
// the project.
func Build() cli.Command {
	return cli.Command{
		Name:  "build",
		Usage: "Builds any modified apps within the project",
		Action: withModAndProject(func(ctx *cli.Context, mod deps.Module, pj *config.ProjectFile) error {
			return command.Build(mod, pj)
		}),
	}
}