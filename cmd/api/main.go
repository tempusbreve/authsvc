package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"

	"breve.us/authsvc"
)

var (
	apiRoot   = "/api/"
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
		Value:  authsvc.Generate(authsvc.HashKeySize),
	}
	seedBlockFlag = cli.StringFlag{
		Name:   "seedblock",
		Usage:  "base64 encoded seed block (default is transient)",
		EnvVar: "SEED_BLOCK",
		Value:  authsvc.Generate(authsvc.BlockKeySize),
	}
	corsOriginsFlag = cli.StringSliceFlag{
		Name:   "cors_origins",
		Usage:  "define CORS acceptable origins (default is insecure!)",
		EnvVar: "CORS_ORIGINS",
		Value:  &cli.StringSlice{"*"},
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
			realmFlag,
			seedHashFlag,
			seedBlockFlag,
			corsOriginsFlag}}}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx *cli.Context) error {
	listen := fmt.Sprintf(":%d", ctx.Int("port"))
	verbose := ctx.Bool("verbose")
	seeder, err := authsvc.NewSeeder(ctx.String("seedhash"), ctx.String("seedblock"))
	if err != nil {
		return err
	}

	l, err := authsvc.NewAuthenticationMiddleware("myrealm", authRoot, seeder)
	if err != nil {
		return err
	}

	sec := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		FrameDeny:          true,

		IsDevelopment: true,
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   ctx.StringSlice("cors_origins"),
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: true,

		Debug: false,
	})

	n := negroni.New(
		negroni.NewRecovery(),
		authsvc.NewDebugMiddleware(os.Stderr, verbose),
		negroni.HandlerFunc(sec.HandlerFuncWithNext),
		negroni.HandlerFunc(c.ServeHTTP),
		negroni.NewStatic(http.Dir(ctx.String("public"))),
	)

	r := mux.NewRouter()

	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(l.LoginHandler())))

	a := n.With(negroni.Handler(l))
	r.PathPrefix(apiRoot).Handler(a.With(negroni.Wrap(authsvc.RegisterAPI(apiRoot, verbose))))
	r.PathPrefix(oauthRoot).Handler(a.With(negroni.Wrap(authsvc.RegisterOAuth(oauthRoot))))

	r.PathPrefix("/").Handler(n.With(negroni.WrapFunc(defaultHandlerFn)))

	s := &http.Server{
		Addr:           listen,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	log.Printf("listening on %s\n", listen)
	return s.ListenAndServe()
}

func defaultHandlerFn(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, " THE END ")
}
