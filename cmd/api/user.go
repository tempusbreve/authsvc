package main

import (
	"fmt"

	"github.com/urfave/cli"

	"breve.us/authsvc/user"
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
		Flags: []cli.Flag{
			ldapHostFlag,
			ldapPortFlag,
			ldapBaseDNFlag,
			ldapUserFlag,
			ldapPasswordFlag,
		},
	}
}

func listUsers(ctx *cli.Context) error {
	cfg := &user.LDAPConfig{
		Host:     ctx.String(ldapHost),
		Port:     ctx.Int(ldapPort),
		Username: ctx.String(ldapUser),
		Password: ctx.String(ldapPass),
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

const (
	ldapHost   = "ldaphost"
	ldapPort   = "ldapport"
	ldapBaseDN = "ldapbase"
	ldapUser   = "ldapuser"
	ldapPass   = "ldappass"
)

var (
	ldapHostFlag = cli.StringFlag{
		Name:   ldapHost,
		Usage:  "hostname for LDAP connection",
		EnvVar: "LDAP_HOST",
		Value:  "localhost",
	}
	ldapPortFlag = cli.IntFlag{
		Name:   ldapPort,
		Usage:  "port for LDAP connection",
		EnvVar: "LDAP_PORT",
		Value:  389,
	}
	ldapBaseDNFlag = cli.StringFlag{
		Name:   ldapBaseDN,
		Usage:  "base DN for LDAP connection",
		EnvVar: "LDAP_BASE_DN",
		Value:  "dc=example,dc=org",
	}
	ldapUserFlag = cli.StringFlag{
		Name:   ldapUser,
		Usage:  "username for LDAP connection",
		EnvVar: "LDAP_USER",
		Value:  "cn=admin,dc=example,dc=org",
	}
	ldapPasswordFlag = cli.StringFlag{
		Name:   ldapPass,
		Usage:  "password for LDAP connection",
		EnvVar: "LDAP_PASS",
		Value:  "",
	}
)
