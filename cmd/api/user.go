package main

import (
	"fmt"

	"github.com/urfave/cli"
)

func newUserCmd() cli.Command {
	return cli.Command{
		Name: "user",
		Subcommands: cli.Commands{
			newListUsersCmd(),
		},
	}
}

func newListUsersCmd() cli.Command {
	return cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Action:  listUsers,
	}
}

func listUsers(ctx *cli.Context) error {
	fmt.Fprintf(ctx.App.Writer, "Hello\n")
	return nil
}
