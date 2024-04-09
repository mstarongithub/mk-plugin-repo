package server

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/volatiletech/authboss/v3"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type Server struct {
	storage    *storage.Storage
	router     http.Handler
	frontendFS fs.FS
}

func NewServer(frontendFS fs.FS, indexFile string, ab *authboss.Authboss) (*Server, error) {
	mainRouter := http.NewServeMux()

	frontendRouter, _ := buildFrontendRouter(frontendFS, indexFile)
	apiRouter, _ := buildApiRouter(ab)

	mainRouter.Handle("/", frontendRouter)
	mainRouter.Handle("/api", apiRouter)

	server := Server{
		storage:    nil,
		router:     frontendRouter,
		frontendFS: frontendFS,
	}
	return &server, nil
}

func buildFrontendRouter(frontendFS fs.FS, index string) (http.Handler, error) {
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, index)
	})

	router.HandleFunc("GET /_app/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServerFS(frontendFS).ServeHTTP(w, r)
	})
	return router, nil
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func buildApiRouter(ab *authboss.Authboss) (http.Handler, error) {
	router := http.NewServeMux()

	router.Handle("/auth", ab.Core.Router)
	return router, nil
}
