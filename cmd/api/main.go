package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"breve.us/authsvc"

	"github.com/urfave/negroni"
)

var (
	listen  = ":4884"
	verbose = false
	public  = "public"
)

func main() {
	if p := os.Getenv("PORT"); p != "" {
		listen = ":" + p
	}
	if v := os.Getenv("VERBOSE"); v != "" {
		log.Printf("Verbose Logging enabled")
		verbose = true
	}
	if p := os.Getenv("PUBLIC_DIR"); p != "" {
		public = p
	}

	n := negroni.New(
		negroni.NewRecovery(),
		authsvc.DebugLogger(os.Stderr, verbose),
		negroni.NewStatic(http.Dir(public)),
		negroni.Wrap(authsvc.RegisterAPI(verbose)),
	)

	s := &http.Server{
		Addr:           listen,
		Handler:        n,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	log.Printf("listening on %s\n", listen)
	log.Fatal(s.ListenAndServe())
}
