package cmd // import "breve.us/authsvc/cmd"

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
)

func newUserCmd() cli.Command {
	return cli.Command{
		Name: "user",
		Subcommands: cli.Commands{
			newGetUserCmd(),
			newCheckPasswordCmd(),
			newListUsersCmd(),
		},
	}
}

func newGetUserCmd() cli.Command {
	return cli.Command{
		Name:    "get",
		Aliases: []string{"g"},
		Action:  getUser,
		Flags: []cli.Flag{
			ldapHostFlag,
			ldapPortFlag,
			ldapBaseDNFlag,
			ldapAdminUserFlag,
			ldapAdminPassFlag,
		},
	}
}

func getUser(ctx *cli.Context) error {
	cache := user.NewLDAPCache(&store.LDAPConfig{
		Host:     ctx.String(ldapHost),
		Port:     ctx.Int(ldapPort),
		Username: ctx.String(ldapAdminUser),
		Password: ctx.String(ldapAdminPass),
		BaseDN:   ctx.String(ldapBaseDN),
	})
	uid := ctx.Args().First()
	if uid == "" {
		return errors.New("expecting uid as parameter")
	}
	m, err := cache.Get(uid)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(ctx.App.Writer, "%+v\n", m)
	return err
}

func newCheckPasswordCmd() cli.Command {
	return cli.Command{
		Name:    "check",
		Aliases: []string{"pwd", "c"},
		Action:  checkPassword,
		Flags: []cli.Flag{
			ldapHostFlag,
			ldapPortFlag,
			ldapBaseDNFlag,
			ldapAdminUserFlag,
			ldapAdminPassFlag,
		},
	}
}

func checkPassword(ctx *cli.Context) error {
	checker := user.NewLDAPChecker(&store.LDAPConfig{
		Host:     ctx.String(ldapHost),
		Port:     ctx.Int(ldapPort),
		Username: ctx.String(ldapAdminUser),
		Password: ctx.String(ldapAdminPass),
		BaseDN:   ctx.String(ldapBaseDN),
	})
	a := ctx.Args()
	if ctx.NArg() < 2 {
		return errors.New("expecting uid and pwd as parameters")
	}

	var err error

	uid := a.Get(0)
	pwd := a.Get(1)
	if checker.IsAuthenticated(uid, pwd) {
		_, err = fmt.Fprintf(ctx.App.Writer, "Authenticated!\n")
		return err
	}
	_, err = fmt.Fprintf(ctx.App.ErrWriter, "Not Authenticated\n")
	return err
}

func newListUsersCmd() cli.Command {
	return cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Action:  listUsers,
		Flags: []cli.Flag{
			ldapHostFlag,
			ldapPortFlag,
			ldapBaseDNFlag,
			ldapAdminUserFlag,
			ldapAdminPassFlag,
		},
	}
}

func listUsers(ctx *cli.Context) error {
	cfg := &store.LDAPConfig{
		Host:     ctx.String(ldapHost),
		Port:     ctx.Int(ldapPort),
		Username: ctx.String(ldapAdminUser),
		Password: ctx.String(ldapAdminPass),
		BaseDN:   ctx.String(ldapBaseDN),
	}
	keys, err := user.NewLDAPCache(cfg).Keys()
	if err != nil {
		return err
	}
	for _, k := range keys {
		fmt.Fprintf(ctx.App.Writer, "%s\n", k)
	}
	return nil
}
