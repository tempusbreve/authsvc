package cmd // import "breve.us/authsvc/cmd"

import (
	"fmt"
	"log"
	"net/http"
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

// NewAuthServerApp creates an App for cli usage
func NewAuthServerApp(version string) *cli.App {
	app := cli.NewApp()
	app.Usage = "authorization server"
	app.Version = version
	app.Action = authServer
	app.Flags = []cli.Flag{
		portFlag,
		bindFlag,
		debugFlag,
		corsOriginsFlag,
		insecureFlag,
		publicHomeFlag,
		hashFlag,
		blockFlag,
		cacheDirFlag,
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

func authServer(ctx *cli.Context) error {
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

	userRegistry := user.NewRegistry(user.NewLDAPCache(ldapCfg))

	var (
		err          error
		provider     common.KeyProvider
		oauthHandler *authorization.OAuthHandler
	)

	if provider, err = common.NewKeyProvider(ctx.String(crypthash), ctx.String(cryptblock)); err != nil {
		return err
	}

	checker := authentication.NewSecureCookieChecker(provider, userRegistry)
	exposedRoutes := []string{}
	options := &authentication.Options{
		Realm:          realm,
		PublicRoots:    exposedRoutes,
		LoginPath:      ctx.String(loginPath),
		RequestChecker: checker,
	}

	if oauthHandler, err = authorization.NewHandler(&authorization.Options{CacheDir: ctx.String(cacheDir), Users: userRegistry}); err != nil {
		return err
	}

	pchecker := common.PasswordCheckers(user.NewLDAPChecker(ldapCfg))

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

	n := negroni.New(
		negroni.NewRecovery(),
		common.NewDebugMiddleware(ctx.App.ErrWriter, ctx.Bool(debug)),
		negroni.HandlerFunc(sec.HandlerFuncWithNext),
		negroni.HandlerFunc(c.ServeHTTP),
		negroni.NewStatic(http.Dir(ctx.String(publicHome))),
	)

	r := mux.NewRouter()

	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(authentication.LoginHandler(authRoot, pchecker, provider, ctx.Bool(insecure)))))

	a := n.With(authentication.NewMiddleware(authRoot, options))
	r.PathPrefix(oauthRoot).Handler(a.With(negroni.Wrap(oauthHandler.RegisterAPI(oauthRoot))))

	r.PathPrefix("/").Handler(n.With(negroni.WrapFunc(defaultHandlerFn)))

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ctx.String(listenIP), ctx.Int(listenPort)),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	log.Printf("%v version %v", ctx.App.Name, ctx.App.Version)
	log.Printf("listening on %s\n", s.Addr)
	return s.ListenAndServe()
}

func defaultHandlerFn(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprintf(w, "no handler"); err != nil {
		log.Printf("error writing response: %v", err)
	}
}
