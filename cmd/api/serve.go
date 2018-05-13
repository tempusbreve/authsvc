package main

import (
	"encoding/json"
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

	"breve.us/authsvc/authentication"
	"breve.us/authsvc/authorization"
	"breve.us/authsvc/client"
	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
)

const (
	listenPort     = "port"
	verboseLog     = "verbose"
	publicHome     = "public"
	dataHome       = "data"
	storageEngine  = "storage"
	basicPasswords = "passwords"
	clients        = "clients"
	users          = "users"
	realm          = "realm"
	seedhash       = "seedhash"
	seedblock      = "seedblock"
	corsOrigins    = "cors_origins"
	insecure       = "insecure"

	userRoot  = "/api/v4/user"
	authRoot  = "/auth/"
	oauthRoot = "/oauth/"
)

var (
	portFlag = cli.IntFlag{
		Name:   listenPort,
		Usage:  "api listen port",
		EnvVar: "PORT",
		Value:  4884,
	}
	verboseFlag = cli.BoolFlag{
		Name:   verboseLog,
		Usage:  "increase logging level",
		EnvVar: "VERBOSE",
	}
	publicFlag = cli.StringFlag{
		Name:   publicHome,
		Usage:  "path to public folder",
		EnvVar: "PUBLIC_HOME",
		Value:  "public",
	}
	storageFlag = cli.StringFlag{
		Name:   storageEngine,
		Usage:  "storage engine [memory or boltdb]",
		EnvVar: "STORAGE_ENGINE",
		Value:  "boltdb",
	}
	dataFlag = cli.StringFlag{
		Name:   dataHome,
		Usage:  "path to data folder",
		EnvVar: "DATA_HOME",
		Value:  "data",
	}
	passwordsFlag = cli.StringFlag{
		Name:   basicPasswords,
		Usage:  "JSON-encoded map of usernames to bcrypted passwords",
		EnvVar: "PASSWORDS",
		Value:  "",
	}
	clientsFlag = cli.StringFlag{
		Name:   clients,
		Usage:  "path to JSON-encoded registered OAuth2 clients",
		EnvVar: "CLIENTS",
		Value:  "",
	}
	usersFlag = cli.StringFlag{
		Name:   users,
		Usage:  "path to JSON-encoded registered users",
		EnvVar: "USERS",
		Value:  "",
	}
	realmFlag = cli.StringFlag{
		Name:   realm,
		Usage:  "authentication realm",
		EnvVar: "REALM",
		Value:  "authsvc",
	}
	seedHashFlag = cli.StringFlag{
		Name:   seedhash,
		Usage:  "base64 encoded seed hash (default is transient)",
		EnvVar: "SEED_HASH",
		Value:  common.Generate(common.HashKeySize),
	}
	seedBlockFlag = cli.StringFlag{
		Name:   seedblock,
		Usage:  "base64 encoded seed block (default is transient)",
		EnvVar: "SEED_BLOCK",
		Value:  common.Generate(common.BlockKeySize),
	}
	corsOriginsFlag = cli.StringSliceFlag{
		Name:   corsOrigins,
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

func newServeCmd() cli.Command {
	return cli.Command{
		Name:   "serve",
		Action: serve,
		Flags: []cli.Flag{
			storageFlag,
			dataFlag,
			passwordsFlag,
			clientsFlag,
			usersFlag,
			publicFlag,
			realmFlag,
			portFlag,
			corsOriginsFlag,
			insecureFlag,
			seedHashFlag,
			seedBlockFlag,
			verboseFlag,
		}}
}

func serve(ctx *cli.Context) error {
	listen := fmt.Sprintf(":%d", ctx.Int(listenPort))
	verbose := ctx.Bool(verboseLog)

	oopts, err := buildOAuthOptions(ctx)
	if err != nil {
		return err
	}

	authMiddleware, oauthHandler, err := buildAuth(ctx, oopts)
	if err != nil {
		return err
	}

	uopts := user.Options{
		Root:    userRoot,
		Verbose: verbose,
		Users:   oopts.Users,
	}

	sec := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		FrameDeny:          true,

		IsDevelopment: true,
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   ctx.StringSlice(corsOrigins),
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: true,

		Debug: false,
	})

	n := negroni.New(
		negroni.NewRecovery(),
		common.NewDebugMiddleware(os.Stderr, verbose),
		negroni.HandlerFunc(sec.HandlerFuncWithNext),
		negroni.HandlerFunc(c.ServeHTTP),
		negroni.NewStatic(http.Dir(ctx.String(publicHome))),
	)

	r := mux.NewRouter()

	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(authMiddleware.LoginHandler())))

	a := n.With(negroni.Handler(authMiddleware))
	r.PathPrefix(userRoot).Handler(a.With(negroni.Wrap(user.RegisterAPI(uopts))))
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
	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprintf(w, "no handler"); err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func buildAuth(ctx *cli.Context, opts *authorization.Options) (*authentication.Middleware, *authorization.Handler, error) {
	oh := authorization.NewHandler(opts)

	var checker common.PasswordChecker
	if opts.Users != nil {
		checker = common.PasswordCheckers(opts.Users.PlainTextChecker(), opts.Users.BcryptChecker())
	}

	var requestChecker common.RequestChecker
	if opts.TokenCache != nil && opts.Users != nil {
		requestChecker = authorization.NewRequestChecker(opts.TokenCache, opts.Users)
	}
	ao, err := buildAuthOptions(ctx, checker, requestChecker, opts.Users)
	if err != nil {
		return nil, nil, err
	}

	am, err := authentication.NewMiddleware(authRoot, ao)
	if err != nil {
		return nil, nil, err
	}
	return am, oh, nil
}

type caches struct {
	stateCache   store.Cache
	clientCache  store.Cache
	userCache    store.Cache
	clientTokens store.Cache
	tokenClients store.Cache
}

func createBoltStorage(data string) (*caches, error) {
	var (
		sc, uc, ct, tc, cc store.Cache
		err                error
	)
	if sc, err = store.NewBoltDBCache(data, "cache"); err != nil {
		return nil, err
	}
	if cc, err = store.NewBoltDBCache(data, "clients"); err != nil {
		return nil, err
	}
	if uc, err = store.NewBoltDBCache(data, "users"); err != nil {
		return nil, err
	}
	if ct, err = store.NewBoltDBCache(data, "clienttokens"); err != nil {
		return nil, err
	}
	if tc, err = store.NewBoltDBCache(data, "tokenclients"); err != nil {
		return nil, err
	}
	return &caches{
		stateCache:   sc,
		clientCache:  cc,
		userCache:    uc,
		clientTokens: ct,
		tokenClients: tc,
	}, nil
}

func createMemoryStorage() *caches {
	return &caches{
		stateCache:   store.NewMemoryCache(),
		clientCache:  store.NewMemoryCache(),
		userCache:    store.NewMemoryCache(),
		clientTokens: store.NewMemoryCache(),
		tokenClients: store.NewMemoryCache(),
	}
}

func buildOAuthOptions(ctx *cli.Context) (*authorization.Options, error) {
	var (
		tok *authorization.TokenCache
		cr  *client.Registry
		ur  *user.Registry
		c   *caches
		err error
	)

	switch ctx.String(storageEngine) {
	case "boltdb":
		data := path.Join(ctx.String(dataHome), "storage.db")
		if c, err = createBoltStorage(data); err != nil {
			return nil, err
		}
	default:
		c = createMemoryStorage()
	}

	tok = authorization.NewTokenCache(c.clientTokens, c.tokenClients)
	cr = client.NewRegistry(c.clientCache)
	ur = user.NewRegistry(c.userCache)

	if clientsFile := ctx.String(clients); clientsFile != "" {
		var rdr io.ReadCloser
		if rdr, err = os.Open(clientsFile); err == nil {
			err = cr.LoadFromJSON(rdr)
		}
		if err != nil {
			return nil, err
		}
	}

	if usersFile := ctx.String(users); usersFile != "" {
		var rdr io.ReadCloser
		if rdr, err = os.Open(usersFile); err == nil {
			err = ur.LoadFromJSON(rdr)
		}
		if err != nil {
			return nil, err
		}
	}

	return &authorization.Options{
		Cache:      c.stateCache,
		TokenCache: tok,
		Clients:    cr,
		Users:      ur,
	}, nil
}

func buildAuthOptions(
	ctx *cli.Context,
	checker common.PasswordChecker,
	requestChecker common.RequestChecker,
	users *user.Registry) (*authentication.Options, error) {

	seeder, err := common.NewKeyProvider(ctx.String(seedhash), ctx.String(seedblock))
	if err != nil {
		return nil, err
	}

	if pwdFile := ctx.String(basicPasswords); pwdFile != "" {
		var rdr io.ReadCloser
		if rdr, err = os.Open(pwdFile); err != nil {
			return nil, err
		}
		defer func() { _ = rdr.Close() }()

		var creds map[string]string
		if err = json.NewDecoder(rdr).Decode(&creds); err != nil {
			return nil, err
		}
		checker = common.PasswordCheckers(authentication.NewBasicChecker(creds), checker)
	}

	if checker == nil {
		checker = authentication.NewBasicChecker(nil)
	}

	return authentication.NewConfig(
		ctx.String(realm),
		[]string{path.Join(oauthRoot, "token")},
		ctx.Bool(insecure),
		seeder,
		checker,
		requestChecker,
		users)
}
