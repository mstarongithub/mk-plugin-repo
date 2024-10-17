package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/mstarongithub/passkey"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type Server struct {
	storage    *storage.Storage
	handler    http.Handler
	frontendFS fs.FS
}

type ServerContextKey string

const (
	CONTEXT_KEY_SERVER     = ServerContextKey("server")
	CONTEXT_KEY_STORAGE    = ServerContextKey("storage")
	CONTEXT_KEY_CSRF_TOKEN = ServerContextKey("csrf_token")
	CONTEXT_KEY_AUTH_LAYER = ServerContextKey("auth-layer")
	CONTEXT_KEY_LOG        = ServerContextKey("logging")
	CONTEXT_KEY_ACTOR_ID   = ServerContextKey("actor-id")
	CONTEXT_KEY_ACTOR_NAME = "actor-name"
)

func NewServer(
	frontendFS fs.FS,
	store *storage.Storage,
	pkey *passkey.Passkey,
) (*Server, error) {
	mainRouter := http.NewServeMux()

	frontendRouter := buildFrontendRouter(frontendFS)
	apiRouter := buildApiRouter(pkey)

	mainRouter.Handle("/", frontendRouter)
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mainRouter.HandleFunc(
		"/alive",
		func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "meow") },
	)
	pkey.MountRoutes(mainRouter, "/webauthn/")

	server := Server{
		storage:    store,
		handler:    nil,
		frontendFS: frontendFS,
	}

	server.handler = ChainMiddlewares(
		mainRouter,
		ContextValsMiddleware(
			map[any]any{
				CONTEXT_KEY_SERVER:  &server,
				CONTEXT_KEY_STORAGE: store,
				// CONTEXT_KEY_AUTH_LAYER: authLayer,
			},
		),
		cors.AllowAll().Handler,
		WebLoggerWrapper,
	)

	return &server, nil
}

func buildFrontendRouter(frontendFS fs.FS) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, frontendFS, "index.html")
	})

	router.HandleFunc("GET /{file}", func(w http.ResponseWriter, r *http.Request) {
		fileName := r.PathValue("file")
		if len(strings.Split(fileName, ".")) == 1 {
			fileName += ".html"
		}
		http.ServeFileFS(w, r, frontendFS, fileName)
	})

	router.HandleFunc("GET /_app/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServerFS(frontendFS).ServeHTTP(w, r)
	})
	return router
}

func (s *Server) Run(addr string) error {
	log.Info().Str("address", addr).Msg("Starting webserver")
	return http.ListenAndServe(addr, s.handler)
}

func buildApiRouter(pkey *passkey.Passkey) http.Handler {
	router := http.NewServeMux()

	router.Handle("/v1/", http.StripPrefix("/v1", buildV1Router(pkey)))
	router.Handle("/metrics/", http.StripPrefix("/metrics", buildMetricsHandler()))
	return router
}

func buildV1Router(pkey *passkey.Passkey) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /plugins", getPluginList)
	router.HandleFunc("GET /plugins/{pluginId}", getSpecificPlugin)
	router.HandleFunc("GET /plugins/{pluginId}/{versionName}", getVersion)

	// router.HandleFunc("GET /auth/login/start", AuthLoginPWHandler)
	// router.HandleFunc("POST /auth/login/mfa", AuthLoginMfaHandler)

	// router.HandleFunc("POST /auth/register/start", authRegisterStartHandler)
	// router.HandleFunc("POST /auth/register/password", authRegisterAddPasswordHandler)
	// router.HandleFunc("POST /auth/register/mail", authRegisterAddMailHandler)
	// router.HandleFunc("POST /auth/register/description", authRegisterAddDescriptionHandler)
	// router.HandleFunc("POST /auth/register/finalise", authRegisterFinaliseHandler)
	// router.HandleFunc("POST /auth/register/cancel", authRegisterCancelHandler)
	router.Handle("/", buildV1RestrictedRouter(pkey))

	return router
}

func buildV1RestrictedRouter(pkey *passkey.Passkey) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/forbidden-test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello there")
	})

	router.HandleFunc("POST /plugins", addNewPlugin)
	router.HandleFunc("PUT /plugins/{pluginId}", updateSpecificPlugin)
	router.HandleFunc("POST /plugins/{pluginId}", newVersion)
	router.HandleFunc("DELETE /plugins/{pluginId}", deleteSpecificPlugin)
	router.HandleFunc("DELETE /plugins/{pluginId}/{versionName}", hideVersion)
	router.HandleFunc("GET /tokens", GetAllTokens)
	router.HandleFunc("PUT /tokens/{tokenName}", GenerateNewToken)
	router.HandleFunc("POST /tokens", ExtendToken)
	router.HandleFunc("DELETE /tokens/{name}", InvalidateToken)

	router.Handle(
		"/admin/users/",
		http.StripPrefix("/admin/users", buildV1AccountAdminRouter()),
	)
	router.Handle(
		"/admin/plugins/",
		http.StripPrefix("/admin/plugins", buildV1PluginAdminRouter()),
	)
	router.HandleFunc("POST /delete", DeleteAccountHandler)

	return pkey.Auth(
		CONTEXT_KEY_ACTOR_NAME,
		nil,
		// TODO: Replace this with func to also check for token before redirecting or refusing
		passkey.RedirectUnauthorized(url.URL{Path: "/"}),
	)(
		ChainMiddlewares(router, passkeyAuthInsertUidMiddleware),
	)
}

func buildV1AccountAdminRouter() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("POST /approve", VerifyUserHandler)
	router.HandleFunc("GET /unapproved", GetAllUnverifiedAccountsHandler)
	router.HandleFunc("POST /promote-admin/plugins", PromotePluginAdminHandler)
	router.HandleFunc("POST /promote-admin/accounts", PromoteAccountAdminHandler)
	router.HandleFunc("POST /demote-admin/plugins", DemotePluginAdminHandler)
	router.HandleFunc("POST /demote-admin/accounts", DemoteAccountAdminHandler)
	router.HandleFunc("GET /userdata/{id}", InspectAccountAdminHandler)

	var handler http.Handler
	handler = ChainMiddlewares(router, CanApproveUsersOnlyMiddleware)

	return handler
}

func buildV1PluginAdminRouter() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("POST /approve", VerifyNewPluginHandler)
	router.HandleFunc("GET /unapproved", GetAllUnverifiedPluginshandler)

	var handler http.Handler
	handler = ChainMiddlewares(router, CanApproveNotesOnlyMiddleware)
	return handler
}
