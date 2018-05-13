package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

var version = "0.0.1"

func main() {
	app := cli.NewApp()
	app.Usage = "api cli"
	app.Version = version
	app.Commands = []cli.Command{
		newServeCmd(),
		newGenerateCmd(),
		newBcryptCmd(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
