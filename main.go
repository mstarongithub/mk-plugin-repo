package main

import (
	"embed"
	"flag"

	"github.com/sirupsen/logrus"
	_ "github.com/volatiletech/authboss-renderer"
	abrenderer "github.com/volatiletech/authboss-renderer"

	_ "github.com/mstarongithub/mk-plugin-repo/auth-old"
	authold "github.com/mstarongithub/mk-plugin-repo/auth-old"
	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/fswrapper"
	"github.com/mstarongithub/mk-plugin-repo/server"
	"github.com/mstarongithub/mk-plugin-repo/storage"
	"github.com/mstarongithub/mk-plugin-repo/util"
)

const DB_DEFAULT_FILE = "db.sqlite"

//go:embed frontend/build frontend/build/_app/*
var frontendFS embed.FS

var level = flag.String(
	"loglevel",
	"info",
	"Set the log level of the app to one of \"debug\", \"info\", \"warn\", \"error\"",
)

func main() {
	setLogLevelFromArgs()
	_, err := config.ReadConfig(nil)
	if err != nil {
		logrus.WithError(err).Warnln("Err reading config, using default")
		config.SetGlobalToDefault()
	}

	err = util.CreateFileIfNotExists(DB_DEFAULT_FILE)
	if err != nil {
		panic(err)
	}

	store, err := storage.NewStorage(DB_DEFAULT_FILE, nil)
	if err != nil {
		panic(err)
	}
	ab, err := authold.SetupAuthboss(
		&store,
		[]byte("placeholder string"),
		[]byte("placeholder string"),
		abrenderer.NewEmail("/auth", "ab_views"),
	)
	if err != nil {
		panic(err)
	}
	httpServer, err := server.NewServer(
		fswrapper.NewFSWrapper(frontendFS, "frontend/build/", false),
		ab,
		&store,
	)
	if err != nil {
		panic(err)
	}
	panic(httpServer.Run(":8080"))
}

func setLogLevelFromArgs() {
	flag.Parse()
	switch *level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.WarnLevel)
	}
}
