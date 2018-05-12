package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"breve.us/authsvc/common"
)

var (
	userRoot  = "/api/v4/user"
	authRoot  = "/auth/"
	oauthRoot = "/oauth/"
	version   = "0.0.1"

	portFlag = cli.IntFlag{
		Name:   "port",
		Usage:  "api listen port",
		EnvVar: "PORT",
		Value:  4884,
	}
	verboseFlag = cli.BoolFlag{
		Name:   "verbose",
		Usage:  "increase logging level",
		EnvVar: "VERBOSE",
	}
	publicFlag = cli.StringFlag{
		Name:   "public",
		Usage:  "path to public folder",
		EnvVar: "PUBLIC_HOME",
		Value:  "public",
	}
	storageFlag = cli.StringFlag{
		Name:   "storage",
		Usage:  "storage engine [memory or boltdb]",
		EnvVar: "STORAGE_ENGINE",
		Value:  "boltdb",
	}
	dataFlag = cli.StringFlag{
		Name:   "data",
		Usage:  "path to data folder",
		EnvVar: "DATA_HOME",
		Value:  "data",
	}
	passwordsFlag = cli.StringFlag{
		Name:   "passwords",
		Usage:  "JSON-encoded map of usernames to bcrypted passwords",
		EnvVar: "PASSWORDS",
		Value:  "",
	}
	clientsFlag = cli.StringFlag{
		Name:   "clients",
		Usage:  "path to JSON-encoded registered OAuth2 clients",
		EnvVar: "CLIENTS",
		Value:  "",
	}
	realmFlag = cli.StringFlag{
		Name:   "realm",
		Usage:  "authentication realm",
		EnvVar: "REALM",
		Value:  "authsvc",
	}
	seedHashFlag = cli.StringFlag{
		Name:   "seedhash",
		Usage:  "base64 encoded seed hash (default is transient)",
		EnvVar: "SEED_HASH",
		Value:  common.Generate(common.HashKeySize),
	}
	seedBlockFlag = cli.StringFlag{
		Name:   "seedblock",
		Usage:  "base64 encoded seed block (default is transient)",
		EnvVar: "SEED_BLOCK",
		Value:  common.Generate(common.BlockKeySize),
	}
	corsOriginsFlag = cli.StringSliceFlag{
		Name:   "cors_origins",
		Usage:  "define CORS acceptable origins (default is insecure!)",
		EnvVar: "CORS_ORIGINS",
		Value:  &cli.StringSlice{"*"},
	}
	insecureFlag = cli.BoolFlag{
		Name:   "insecure",
		Usage:  "don't use secure cookie",
		EnvVar: "INSECURE",
	}
)

func main() {
	app := cli.NewApp()
	app.Usage = "api cli"
	app.Version = version
	app.Commands = []cli.Command{{
		Name:   "serve",
		Action: serve,
		Flags: []cli.Flag{
			portFlag,
			verboseFlag,
			publicFlag,
			storageFlag,
			dataFlag,
			passwordsFlag,
			clientsFlag,
			realmFlag,
			seedHashFlag,
			seedBlockFlag,
			corsOriginsFlag,
			insecureFlag}}, {
		Name:   "bcrypt",
		Action: crypt,
	}}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
