package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"

	"breve.us/authsvc/client"
	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
)

func newGenerateCmd() cli.Command {
	return cli.Command{
		Name:    "generate",
		Aliases: []string{"gen", "g"},
		Subcommands: cli.Commands{
			newUsersCmd(),
			newClientsCmd(),
			newPasswordsCmd(),
		},
	}
}

var (
	outputFileFlag = cli.StringFlag{
		Name:  "out_file",
		Usage: "file to write JSON-encoded generated content",
		Value: "out.json",
	}
	usernameFlag = cli.StringFlag{
		Name:  "username",
		Usage: "username to use",
		Value: "test.user@example.com",
	}
	passwordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "password to use",
		Value: "T0p53cr37",
	}
)

//
// Users
//

func newUsersCmd() cli.Command {
	return cli.Command{
		Name:    "users",
		Aliases: []string{"u"},
		Flags:   []cli.Flag{usernameFlag, passwordFlag, outputFileFlag},
		Action:  generateUsers,
	}
}

func generateUsers(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")
	outFile := ctx.String("out_file")
	if username == "" || password == "" || outFile == "" {
		return errors.New("parameters username, password, and out_file are all required")
	}

	var (
		fd  io.WriteCloser
		pwd string
		err error
	)

	if pwd, err = bcryptFn(password); err != nil {
		return err
	}

	r := user.NewRegistry(store.NewMemoryCache())
	err = r.Put(&user.Details{
		ID:       99,
		Username: username,
		Password: pwd,
		Email:    "user@example.com",
		Name:     "Example User",
		State:    "active",
	})
	if err != nil {
		return err
	}

	if fd, err = open(outFile); err != nil {
		return err
	}
	defer func() { _ = fd.Close() }()

	fmt.Fprintf(ctx.App.Writer, "Writing users file %q, with username %q and password %q\n", outFile, username, password)
	return r.SaveToJSON(fd)
}

//
// Clients
//

func newClientsCmd() cli.Command {
	return cli.Command{
		Name:    "clients",
		Aliases: []string{"cl", "c"},
		Flags:   []cli.Flag{outputFileFlag},
		Action:  generateClients,
	}
}

func generateClients(ctx *cli.Context) error {
	outFile := ctx.String("out_file")
	if outFile == "" {
		return errors.New("parameter out_file is required")
	}

	var (
		fd  io.WriteCloser
		err error
	)

	r := client.NewRegistry(store.NewMemoryCache())
	err = r.Put(&client.Details{
		ID:   "example.com",
		Name: "Sample OAuth2 Client",
		Endpoints: []string{
			"https://example.com/oauth/done",
			"https://example.com/signup/gitlab/complete",
		},
	})
	if err != nil {
		return err
	}

	if fd, err = open(outFile); err != nil {
		return err
	}
	defer func() { _ = fd.Close() }()

	fmt.Fprintf(ctx.App.Writer, "Writing clients file %q\n", outFile)
	return r.SaveToJSON(fd)
}

//
// Passwords
//

func newPasswordsCmd() cli.Command {
	return cli.Command{
		Name:    "passwords",
		Aliases: []string{"pwd", "p"},
		Flags:   []cli.Flag{usernameFlag, passwordFlag, outputFileFlag},
		Action:  generatePasswords,
	}
}

func generatePasswords(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")
	outFile := ctx.String("out_file")
	if username == "" || password == "" || outFile == "" {
		return errors.New("parameters username, password, and out_file are all required")
	}

	var (
		fd  io.WriteCloser
		bc  string
		err error
	)
	if bc, err = bcryptFn(password); err != nil {
		return err
	}
	if fd, err = open(outFile); err != nil {
		return err
	}
	defer func() { _ = fd.Close() }()

	fmt.Fprintf(ctx.App.Writer, "Writing password file %q, with username %q and password %q\n", outFile, username, password)
	enc := json.NewEncoder(fd)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]string{username: bc})
}

func open(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
}
