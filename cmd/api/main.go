package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"breve.us/authsvc"

	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
)

var (
	listen  = ":4884"
	verbose = false
	public  = "public"

	assetRoot = "/assets/"
	apiRoot   = "/api/"
	authRoot  = "/auth/"
	oauthRoot = "/oauth/"
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

	sec := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		FrameDeny:          true,

		IsDevelopment: true,
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // TODO: introduce configuration
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: true,

		Debug: false,
	})

	n := negroni.New(
		negroni.NewRecovery(),
		authsvc.DebugLogger(os.Stderr, verbose),
		negroni.HandlerFunc(sec.HandlerFuncWithNext),
		negroni.HandlerFunc(c.ServeHTTP),
		negroni.Handler(gzip.Gzip(gzip.DefaultCompression)),
		negroni.NewStatic(http.Dir(public)),
	)

	r := mux.NewRouter()

	l := authsvc.NewAuthenticationMiddleware("myrealm", authRoot, nil)
	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(l.LoginHandler())))

	a := n.With(negroni.Handler(l))
	r.PathPrefix(apiRoot).Handler(a.With(negroni.Wrap(authsvc.RegisterAPI(apiRoot, verbose))))
	r.PathPrefix(oauthRoot).Handler(a.With(negroni.Wrap(authsvc.RegisterOAuth(oauthRoot))))

	r.PathPrefix("/").Handler(n)

	s := &http.Server{
		Addr:           listen,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	log.Printf("listening on %s\n", listen)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
