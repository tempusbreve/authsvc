package cmd // import "breve.us/authsvc/cmd"

import (
	"os"

	"github.com/urfave/cli"
)

// NewAPIApp exposes a command line App
func NewAPIApp(version string) *cli.App {
	app := cli.NewApp()
	app.Usage = "api cli"
	app.Version = version
	app.Writer = os.Stdout
	app.ErrWriter = os.Stderr
	app.Commands = []cli.Command{
		newUserCmd(),
		newGenerateCmd(),
		newBcryptCmd(),
	}
	return app
}
