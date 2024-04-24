package main

import (
	"embed"

	"github.com/sirupsen/logrus"
	_ "github.com/volatiletech/authboss-renderer"

	_ "github.com/mstarongithub/mk-plugin-repo/auth-old"
	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/fswrapper"
	"github.com/mstarongithub/mk-plugin-repo/server"
	"github.com/mstarongithub/mk-plugin-repo/storage"
	"github.com/mstarongithub/mk-plugin-repo/util"
)

const DB_DEFAULT_FILE = "db.sqlite"

//go:embed frontend/build frontend/build/_app/*
var frontendFS embed.FS

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	_, err := config.ReadConfig(nil)
	if err != nil {
		logrus.WithError(err).Warnln("Err reading config")
	}

	err = util.CreateFileIfNotExists(DB_DEFAULT_FILE)
	if err != nil {
		panic(err)
	}

	store, err := storage.NewStorage(DB_DEFAULT_FILE, nil)
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
	httpServer, err := server.NewServer(
		fswrapper.NewFSWrapper(frontendFS, "frontend/build/", false),
		nil,
		&store,
	)
	if err != nil {
		panic(err)
	}
	panic(httpServer.Run(":8080"))
}
