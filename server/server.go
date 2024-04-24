package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
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
	ab *authboss.Authboss,
	store *storage.Storage,
) (*Server, error) {
	mainRouter := http.NewServeMux()

	frontendRouter, _ := buildFrontendRouter(frontendFS)
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
			&weblogger.Config{
				DefaultLogLevel:    weblogger.LOG_LEVEL_DEBUG,
				FailedRequestLevel: weblogger.LOG_LEVEL_WARN,
			},
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
	router.HandleFunc("POST /plugins/{pluginId}", newVersion)
	router.HandleFunc("DELETE /plugins/{pluginId}", deleteSpecificPlugin)
	router.HandleFunc("DELETE /plugins/[pluginId]/{versionName}", hideVersion)

	return router
}
