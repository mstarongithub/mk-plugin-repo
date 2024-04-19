package main

import (
	"embed"

	"github.com/sirupsen/logrus"
	_ "github.com/volatiletech/authboss-renderer"

	_ "github.com/mstarongithub/mk-plugin-repo/auth-old"
	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/server"
	"github.com/mstarongithub/mk-plugin-repo/storage"
)

//go:embed frontend
var frontendFS embed.FS

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	_, err := config.ReadConfig(nil)
	if err != nil {
		logrus.WithError(err).Warnln("Err reading config")
	}
	store, err := storage.NewStorage("db.sqlite", nil)
	if err != nil {
		panic(err)
	}
	// ab, err := authold.SetupAuthboss(
	// 	&store,
	// 	[]byte("placeholder string"),
	// 	[]byte("placeholder string"),
	// 	abrenderer.NewEmail("/auth", "ab_views"),
	// )
	// if err != nil {
	// 	panic(err)
	// }
	httpServer, err := server.NewServer(frontendFS, "placeholder lol", nil, &store)
	if err != nil {
		panic(err)
	}
	panic(httpServer.Run(":8080"))
}
