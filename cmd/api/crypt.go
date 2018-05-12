package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
	"golang.org/x/crypto/bcrypt"
)

func crypt(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("provide at least one password to crypt")
	}
	for _, pwd := range ctx.Args() {
		res, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		fmt.Printf("%q --> %q\n", pwd, string(res))
	}
	return nil
}
