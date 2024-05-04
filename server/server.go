package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-pkgz/auth"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

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
	CONTEXT_KEY_AUTHBOSS   = ServerContextKey("authboss")
	CONTEXT_KEY_CSRF_TOKEN = ServerContextKey("csrf_token")
)

func NewServer(
	frontendFS fs.FS,
	store *storage.Storage,
	authLayer *auth.Service,
) (*Server, error) {
	mainRouter := http.NewServeMux()

	frontendRouter, _ := buildFrontendRouter(frontendFS)
	apiRouter, _ := buildApiRouter(authLayer)
	authRoutes, avatarRoutes := authLayer.Handlers()

	mainRouter.Handle("/", frontendRouter)
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mainRouter.Handle("/auth/", http.StripPrefix("/auth", authRoutes))
	mainRouter.Handle("/avatar/", http.StripPrefix("/avatar/", avatarRoutes))

	server := Server{
		storage:    store,
		handler:    nil,
		frontendFS: frontendFS,
	}

	if authLayer != nil {
		authMiddleware := authLayer.Middleware()
		server.handler = ChainMiddlewares(
			mainRouter,
			ContextValsMiddleware(
				map[any]any{
					CONTEXT_KEY_SERVER:  &server,
					CONTEXT_KEY_STORAGE: store,
				},
			),
			cors.AllowAll().Handler,
			authMiddleware.Trace,
			WebLoggerWrapper,
		)
	} else {
		server.handler = ChainMiddlewares(
			mainRouter,
			ContextValsMiddleware(
				map[any]any{
					CONTEXT_KEY_SERVER:  &server,
					CONTEXT_KEY_STORAGE: store,
				},
			),
			cors.AllowAll().Handler,
			WebLoggerWrapper,
		)
	}

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
func buildApiRouter(authLayer *auth.Service) (http.Handler, error) {
	router := http.NewServeMux()

	router.Handle("/v1/", http.StripPrefix("/v1", buildV1Router(authLayer)))
	return router, nil
}

func buildV1Router(authLayer *auth.Service) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /plugins", getPluginList)
	router.HandleFunc("GET /plugins/{pluginId}", getSpecificPlugin)
	router.HandleFunc("GET /plugins/{pluginId}/{versionName}", getVersion)
	router.Handle("/", buildV1RestrictedRouter(authLayer))

	return router
}

func buildV1RestrictedRouter(authLayer *auth.Service) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /plugins", addNewPlugin)
	router.HandleFunc("PUT /plugins/{pluginId}", updateSpecificPlugin)
	router.HandleFunc("POST /plugins/{pluginId}", newVersion)
	router.HandleFunc("DELETE /plugins/{pluginId}", deleteSpecificPlugin)
	router.HandleFunc("DELETE /plugins/{pluginId}/{versionName}", hideVersion)

	var handler http.Handler
	if authLayer != nil {
		middleware := authLayer.Middleware()
		handler = ChainMiddlewares(
			router,
			middleware.Auth,
		)
	} else {
		handler = router
	}

	return handler
}
