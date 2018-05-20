package main

import (
	"log"
	"os"

	"breve.us/authsvc/cmd"
)

var version = "0.0.1"

func main() {
	app := cmd.NewAuthServerApp(version)
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
