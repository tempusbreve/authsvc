package main

import (
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
	"breve.us/authsvc/common"
	"breve.us/authsvc/oauth"
	"breve.us/authsvc/store"
	"breve.us/authsvc/user"
)

func serve(ctx *cli.Context) error {
	listen := fmt.Sprintf(":%d", ctx.Int("port"))
	verbose := ctx.Bool("verbose")

	authMiddleware, oauthHandler, err := buildAuth(ctx)
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

	r.PathPrefix(authRoot).Handler(n.With(negroni.Wrap(authMiddleware.LoginHandler())))

	a := n.With(negroni.Handler(authMiddleware))
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
	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprintf(w, "no handler"); err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func buildAuth(ctx *cli.Context) (*authentication.Middleware, *oauth.Handler, error) {
	opts, err := buildOAuthOptions(ctx)
	if err != nil {
		return nil, nil, err
	}

	oh := oauth.NewHandler(opts)

	ao, err := buildAuthOptions(ctx, oh)
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
	stateCache     store.Cache
	clientTokens   store.Cache
	tokenClients   store.Cache
	clientRegistry store.Cache
	userRegistry   store.Cache
}

func createBoltStorage(data string) (*caches, error) {
	var (
		cache, ct, tc, cr store.Cache
		err               error
	)
	if cache, err = store.NewBoltDBCache(data, "cache"); err != nil {
		return nil, err
	}
	if ct, err = store.NewBoltDBCache(data, "clienttokens"); err != nil {
		return nil, err
	}
	if tc, err = store.NewBoltDBCache(data, "tokenclients"); err != nil {
		return nil, err
	}
	if cr, err = store.NewBoltDBCache(data, "clientregistry"); err != nil {
		return nil, err
	}
	return &caches{
		stateCache:     cache,
		clientTokens:   ct,
		tokenClients:   tc,
		clientRegistry: cr,
		userRegistry:   nil,
	}, nil
}

func createMemoryStorage() *caches {
	return &caches{
		stateCache:     store.NewMemoryCache(),
		clientTokens:   store.NewMemoryCache(),
		tokenClients:   store.NewMemoryCache(),
		clientRegistry: store.NewMemoryCache(),
		userRegistry:   store.NewMemoryCache(),
	}
}

func buildOAuthOptions(ctx *cli.Context) (*oauth.Options, error) {
	var (
		tok     *oauth.TokenCache
		clients *oauth.ClientRegistry
		c       *caches
		err     error
	)

	switch ctx.String("storage") {
	case "boltdb":
		data := path.Join(ctx.String("data"), "storage.db")
		if c, err = createBoltStorage(data); err != nil {
			return nil, err
		}
	default:
		c = createMemoryStorage()
	}

	tok = oauth.NewTokenCache(c.clientTokens, c.tokenClients)
	clients = oauth.NewClientRegistry(c.clientRegistry)

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
		Cache:      c.stateCache,
		TokenCache: tok,
		Clients:    clients,
	}, nil
}

func buildAuthOptions(ctx *cli.Context, oauthHandler *oauth.Handler) (*authentication.Options, error) {
	seeder, err := common.NewSeeder(ctx.String("seedhash"), ctx.String("seedblock"))
	if err != nil {
		return nil, err
	}

	authOpts := &authentication.Options{
		Realm:       ctx.String("realm"),
		Checker:     authentication.NewBasicChecker(nil),
		Seeder:      seeder,
		PublicRoots: []string{path.Join(oauthRoot, "token")},
		OAuth:       oauthHandler,
		Insecure:    ctx.Bool("insecure"),
	}
	return authOpts, nil
}
