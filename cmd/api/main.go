package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/bcrypt"

	"breve.us/authsvc/authentication"
	"breve.us/authsvc/common"
	"breve.us/authsvc/oauth"
	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
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

func serve(ctx *cli.Context) error {
	listen := fmt.Sprintf(":%d", ctx.Int("port"))
	verbose := ctx.Bool("verbose")

	opts, err := buildOAuthOptions(ctx)
	if err != nil {
		return err
	}
	oauthHandler := oauth.NewHandler(opts)

	seeder, err := common.NewSeeder(ctx.String("seedhash"), ctx.String("seedblock"))
	if err != nil {
		return err
	}

	authOpts := &authentication.Options{
		Realm:       ctx.String("realm"),
		Checker:     authentication.NewChecker(nil),
		Seeder:      seeder,
		PublicRoots: []string{path.Join(oauthRoot, "token")},
		OAuth:       oauthHandler,
		Insecure:    ctx.Bool("insecure"),
	}

	l, err := authentication.NewMiddleware(authRoot, authOpts)
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
		common.NewDebugMiddleware(os.Stderr, verbose),
		negroni.HandlerFunc(sec.HandlerFuncWithNext),
		negroni.HandlerFunc(c.ServeHTTP),
		negroni.NewStatic(http.Dir(ctx.String("public"))),
	)

	r := mux.NewRouter()

	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(l.LoginHandler())))

	a := n.With(negroni.Handler(l))
	r.PathPrefix(userRoot).Handler(a.With(negroni.Wrap(user.RegisterAPI(userRoot, verbose))))
	r.PathPrefix(oauthRoot).Handler(a.With(negroni.Wrap(oauthHandler.RegisterAPI(oauthRoot))))

	r.PathPrefix("/").Handler(n.With(negroni.WrapFunc(defaultHandlerFn)))

	s := &http.Server{
		Addr:           listen,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	log.Printf("%v version %v", ctx.App.Name, ctx.App.Version)
	log.Printf("listening on %s\n", listen)
	return s.ListenAndServe()
}

func defaultHandlerFn(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(w, "no handler"); err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func buildOAuthOptions(ctx *cli.Context) (*oauth.Options, error) {
	var (
		cache, ct, tc, cr store.Cache
		tok               *oauth.TokenCache
		clients           *oauth.ClientRegistry
		err               error
	)

	// TODO: some of this should be in the oauth package

	switch ctx.String("storage") {
	case "boltdb":
		data := ctx.String("data")
		if cache, err = store.NewBoltDBCache(path.Join(data, "cache.db")); err != nil {
			return nil, err
		}
		if ct, err = store.NewBoltDBCache(path.Join(data, "clienttokens.db")); err != nil {
			return nil, err
		}
		if tc, err = store.NewBoltDBCache(path.Join(data, "tokenclients.db")); err != nil {
			return nil, err
		}
		if cr, err = store.NewBoltDBCache(path.Join(data, "clientregistry.db")); err != nil {
			return nil, err
		}
	default:
		cache = store.NewMemoryCache()
		ct = store.NewMemoryCache()
		tc = store.NewMemoryCache()
		cr = store.NewMemoryCache()
	}

	tok = oauth.NewTokenCache(ct, tc)
	clients = oauth.NewClientRegistry(cr)

	if clientsFile := ctx.String("clients"); clientsFile != "" {
		var rdr io.ReadCloser
		if rdr, err = os.Open(clientsFile); err == nil {
			err = clients.LoadFromJSON(rdr)
		}
		if err != nil {
			return nil, err
		}
	}

	return &oauth.Options{
		Cache:      cache,
		TokenCache: tok,
		Clients:    clients,
	}, nil
}
