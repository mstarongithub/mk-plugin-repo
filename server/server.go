package server

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/volatiletech/authboss/v3"
	"gitlab.com/mstarongitlab/weblogger"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type Server struct {
	storage    *storage.Storage
	handler    http.Handler
	frontendFS fs.FS
	authboss   *authboss.Authboss
}

type ServerContextKey string

const (
	CONTEXT_KEY_SERVER   = ServerContextKey("server")
	CONTEXT_KEY_STORAGE  = ServerContextKey("storage")
	CONTEXT_KEY_AUTHBOSS = ServerContextKey("authboss")
)

func NewServer(
	frontendFS fs.FS,
	indexFile string,
	ab *authboss.Authboss,
	store *storage.Storage,
) (*Server, error) {
	mainRouter := http.NewServeMux()

	frontendRouter, _ := buildFrontendRouter(frontendFS, indexFile)
	apiRouter, _ := buildApiRouter(ab)

	mainRouter.Handle("/", frontendRouter)
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiRouter))

	server := Server{
		storage:    store,
		handler:    nil,
		frontendFS: frontendFS,
		authboss:   ab,
	}
	server.handler = addContextValsMiddleware(
		weblogger.LoggingMiddleware(
			mainRouter,
			weblogger.LOGGING_SEND_REQUESTS_DEBUG,
		),
		map[any]any{
			CONTEXT_KEY_SERVER:   &server,
			CONTEXT_KEY_STORAGE:  store,
			CONTEXT_KEY_AUTHBOSS: ab,
		},
	)

	return &server, nil
}

// NOTE: Error return value unused currently and can safely be ignored
func buildFrontendRouter(frontendFS fs.FS, index string) (http.Handler, error) {
	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, index)
	})

	router.HandleFunc("GET /_app/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServerFS(frontendFS).ServeHTTP(w, r)
	})
	return router, nil
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.handler)
}

// NOTE: Error return value unused currently and can safely be ignored
func buildApiRouter(_ *authboss.Authboss) (http.Handler, error) {
	router := http.NewServeMux()

	// router.Handle("/auth", ab.Core.Router)
	router.Handle("/v1/", http.StripPrefix("/v1", buildV1Router()))
	return router, nil
}

func buildV1Router() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /plugins", getPluginList)
	router.HandleFunc("GET /plugins/{pluginId}", getSpecificPlugin)
	router.HandleFunc("GET /plugins/{pluginId}/{versionName}", getVersion)
	router.Handle("/", buildV1RestrictedRouter())

	return router
}

func buildV1RestrictedRouter() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /plugins", addNewPlugin)
	router.HandleFunc("PUT /plugins/{pluginId}", updateSpecificPlugin)
	router.HandleFunc("DELETE /plugins/{pluginId}", deleteSpecificPlugin)

	return router
}
