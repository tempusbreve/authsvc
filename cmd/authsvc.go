package cmd // import "breve.us/authsvc/cmd"

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"

	"breve.us/authsvc/authentication"
	"breve.us/authsvc/authorization"
	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
)

// NewAuthSvcApp creates a command line app
func NewAuthSvcApp(version string) *cli.App {
	app := cli.NewApp()
	app.Usage = "authsvc server"
	app.Version = version
	app.Action = authSvc
	app.Flags = []cli.Flag{
		portFlag,
		bindFlag,
		publicHomeFlag,
		templateHomeFlag,
		cacheDirFlag,
		corsOriginsFlag,
		hashFlag,
		blockFlag,
		debugFlag,
		insecureFlag,
		loginPathFlag,
		ldapHostFlag,
		ldapPortFlag,
		ldapTLSFlag,
		ldapAdminUserFlag,
		ldapAdminPassFlag,
		ldapBaseDNFlag,
	}
	return app
}

func authSvc(ctx *cli.Context) error {
	log.SetOutput(ctx.App.Writer)
	log.SetPrefix(logPrefixAuth)
	log.SetFlags(log.LstdFlags | log.Llongfile)

	ldapCfg := &store.LDAPConfig{
		Host:     ctx.String(ldapHost),
		Port:     ctx.Int(ldapPort),
		UseTLS:   ctx.Bool(ldapTLS),
		Username: ctx.String(ldapAdminUser),
		Password: ctx.String(ldapAdminPass),
		BaseDN:   ctx.String(ldapBaseDN),
	}

	pchecker := common.PasswordCheckers(user.NewLDAPChecker(ldapCfg))

	userRegistry := user.NewRegistry(user.NewLDAPCache(ldapCfg))

	provider, err := common.NewKeyProvider(ctx.String(crypthash), ctx.String(cryptblock))
	if err != nil {
		return err
	}

	oauthHandler, err := authorization.NewHandler(&authorization.Options{CacheDir: ctx.String(cacheDir), Users: userRegistry})
	if err != nil {
		return err
	}

	authenticationMiddleware := authentication.NewMiddleware(&authentication.Options{
		Realm:       realm,
		PublicRoots: []string{"/auth/login", "/oauth/token"},
		LoginPath:   ctx.String(loginPath),
		RequestChecker: common.RequestCheckers(
			oauthHandler,
			authentication.NewSecureCookieChecker(provider, userRegistry)),
	})

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

		Debug: ctx.Bool(debug),
	})

	staticAssets := ctx.String(publicHome)
	staticHandler := common.NewStaticFileHandler(path.Join(staticAssets, "index.html"))
	n := negroni.New(
		negroni.NewRecovery(),
		common.NewDebugMiddleware(ctx.App.ErrWriter, ctx.Bool(debug)),
		negroni.HandlerFunc(sec.HandlerFuncWithNext),
		negroni.HandlerFunc(c.ServeHTTP),
		negroni.NewStatic(http.Dir(staticAssets)),
	)

	var userRoot = "/api/v4/user"
	options := user.Options{
		Root:    userRoot,
		Verbose: false, //TODO
		Users:   userRegistry,
	}
	r := mux.NewRouter()

	loginHandler := authentication.LoginHandler(authRoot, pchecker, provider, ctx.Bool(insecure))
	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(loginHandler)))

	oauthAPIHandler := oauthHandler.RegisterAPI(oauthRoot)
	r.PathPrefix(oauthRoot).Handler(n.With(authenticationMiddleware, negroni.Wrap(oauthAPIHandler)))

	userAPIHandler := user.RegisterAPI(options)
	r.PathPrefix(userRoot).Handler(n.With(authenticationMiddleware, negroni.Wrap(userAPIHandler)))

	r.NewRoute().Handler(n.With(negroni.Wrap(staticHandler)))

	for _, rr := range []*mux.Router{r, loginHandler, oauthAPIHandler, userAPIHandler} {
		if err = rr.Walk(fallbackOn(staticHandler)); err != nil {
			return err
		}
	}

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ctx.String(listenIP), ctx.Int(listenPort)),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	log.Printf("%v version %v", ctx.App.Name, ctx.App.Version)
	log.Printf("listening on %s\nstatic content from %q", s.Addr, staticAssets)
	return s.ListenAndServe()
}

func fallbackOn(h http.Handler) func(*mux.Route, *mux.Router, []*mux.Route) error {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		router.NotFoundHandler = h
		router.MethodNotAllowedHandler = h
		return nil
	}
}
