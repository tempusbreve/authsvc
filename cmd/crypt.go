package cmd // import "breve.us/authsvc/cmd"

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
	"golang.org/x/crypto/bcrypt"
)

func newBcryptCmd() cli.Command {
	return cli.Command{
		Name:   "bcrypt",
		Action: crypt,
	}
}

func crypt(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("provide at least one password to crypt")
	}
	for _, pwd := range ctx.Args() {
		res, err := bcryptFn(pwd)
		if err != nil {
			return err
		}
		fmt.Printf("%q --> %q\n", pwd, res)
	}
	return nil
}

func bcryptFn(password string) (string, error) {
	switch c, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err {
	case nil:
		return string(c), nil
	default:
		return "", err
	}
}
