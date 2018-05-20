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
	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
)

// NewUserServerApp creates a command line app
func NewUserServerApp(version string) *cli.App {
	app := cli.NewApp()
	app.Usage = "authorization server"
	app.Version = version
	app.Action = userServer
	app.Flags = []cli.Flag{
		portFlag,
		bindFlag,
		debugFlag,
		corsOriginsFlag,
		publicHomeFlag,
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

func userServer(ctx *cli.Context) error {
	log.SetOutput(ctx.App.Writer)
	log.SetPrefix(logPrefixUser)
	log.SetFlags(log.LstdFlags | log.Llongfile)

	var (
		exposedRoutes []string              // TODO
		checker       common.RequestChecker //TODO
	)

	options := &authentication.Options{
		Realm:          realm,
		PublicRoots:    exposedRoutes,
		LoginPath:      ctx.String(loginPath),
		RequestChecker: checker,
	}

	_ = &store.LDAPConfig{
		Host:     ctx.String(ldapHost),
		Port:     ctx.Int(ldapPort),
		UseTLS:   ctx.Bool(ldapTLS),
		Username: ctx.String(ldapAdminUser),
		Password: ctx.String(ldapAdminPass),
		BaseDN:   ctx.String(ldapBaseDN),
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
	_ = n.With(negroni.Handler(authentication.NewMiddleware("TODO", options)))
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
