package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"github.com/mstarongithub/mk-plugin-repo/auth"
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
)

func NewServer(
	frontendFS fs.FS,
	store *storage.Storage,
	authLayer *auth.Auth,
) (*Server, error) {
	mainRouter := http.NewServeMux()

	frontendRouter, _ := buildFrontendRouter(frontendFS)
	apiRouter, _ := buildApiRouter(authLayer)

	mainRouter.Handle("/", frontendRouter)
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiRouter))

	server := Server{
		storage:    store,
		handler:    nil,
		frontendFS: frontendFS,
	}

	server.handler = ChainMiddlewares(
		mainRouter,
		ContextValsMiddleware(
			map[any]any{
				CONTEXT_KEY_SERVER:     &server,
				CONTEXT_KEY_STORAGE:    store,
				CONTEXT_KEY_AUTH_LAYER: authLayer,
			},
		),
		cors.AllowAll().Handler,
		RouteBasedLoggingMiddleware,
		WebLoggerWrapper,
	)

	return &server, nil
}

// NOTE: Error return value unused currently and can safely be ignored
func buildFrontendRouter(frontendFS fs.FS) (http.Handler, error) {
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
	return router, nil
}

func (s *Server) Run(addr string) error {
	logrus.WithField("adress", addr).Infoln("Starting webserver")
	return http.ListenAndServe(addr, s.handler)
}

// NOTE: Error return value unused currently and can safely be ignored
func buildApiRouter(authLayer *auth.Auth) (http.Handler, error) {
	router := http.NewServeMux()

	router.Handle("/v1/", http.StripPrefix("/v1", buildV1Router(authLayer)))
	return router, nil
}

func buildV1Router(authLayer *auth.Auth) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /plugins", getPluginList)
	router.HandleFunc("GET /plugins/{pluginId}", getSpecificPlugin)
	router.HandleFunc("GET /plugins/{pluginId}/{versionName}", getVersion)

	router.HandleFunc("/auth/login/start", AuthLoginPWHandler)
	router.HandleFunc("POST /auth/login/mfa", AuthLoginMfaHandler)

	router.HandleFunc("POST /auth/register/start", authRegisterStartHandler)
	router.HandleFunc("POST /auth/register/password", authRegisterAddPasswordHandler)
	router.HandleFunc("POST /auth/register/mail", authRegisterAddMailHandler)
	router.HandleFunc("POST /auth/register/description", authRegisterAddDescriptionHandler)
	router.HandleFunc("POST /auth/register/finalise", authRegisterFinaliseHandler)
	router.HandleFunc("POST /auth/register/cancel", authRegisterCancelHandler)
	router.Handle("/", buildV1RestrictedRouter(authLayer))

	return router
}

func buildV1RestrictedRouter(authLayer *auth.Auth) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /plugins", addNewPlugin)
	router.HandleFunc("PUT /plugins/{pluginId}", updateSpecificPlugin)
	router.HandleFunc("POST /plugins/{pluginId}", newVersion)
	router.HandleFunc("DELETE /plugins/{pluginId}", deleteSpecificPlugin)
	router.HandleFunc("DELETE /plugins/{pluginId}/{versionName}", hideVersion)

	var handler http.Handler
	if authLayer != nil {
		handler = ChainMiddlewares(
			router,
			TokenOrAuthMiddleware,
		)
	} else {
		handler = router
	}

	return handler
}
