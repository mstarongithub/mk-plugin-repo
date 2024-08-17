package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/mstarongithub/passkey"
	"github.com/sirupsen/logrus"

	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/fswrapper"
	"github.com/mstarongithub/mk-plugin-repo/server"
	"github.com/mstarongithub/mk-plugin-repo/storage"
	"github.com/mstarongithub/mk-plugin-repo/util"
)

const DB_DEFAULT_FILE = "db.sqlite"

//go:embed frontend/build frontend/build/_app/*
var frontendFS embed.FS

var authMode string = "prod"

func main() {
	setLogLevelFromEnv()
	_, err := config.ReadConfig(nil)
	if err != nil {
		logrus.WithError(err).
			Warnln("Err reading config, using default and attempting to write it to default location")
		config.SetGlobalToDefault()
		config.WriteDefaultConfigToDefaultLocation()
	}

	err = util.CreateFileIfNotExists(DB_DEFAULT_FILE)
	if err != nil {
		panic(err)
	}

	store, err := storage.NewStorage(DB_DEFAULT_FILE, nil)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to setup storage layer")
		os.Exit(1)
	}
	stopServiceWorkersFunc := store.LaunchMiniServices()
	defer stopServiceWorkersFunc()

	// Setup authentication layer
	// var authLayer *auth.Auth
	// logrus.WithField("mode", authMode).Infoln("Using authmode")
	// switch authMode {
	// case "prod":
	// 	authLayer, err = auth.NewAuth(store, auth.AUTH_MODE_DEFAULT)
	// case "dev":
	// 	authLayer, err = auth.NewAuth(store, auth.AUTH_MODE_DEV)
	// case "none":
	// 	authLayer, err = auth.NewAuth(store, auth.AUTH_MODE_NONE)
	// default:
	// 	authLayer, err = auth.NewAuth(store, auth.AUTH_MODE_DEFAULT)
	// }
	// if err != nil {
	// 	logrus.WithError(err).Fatal("Failed to set up authentication layer")
	// 	os.Exit(1)
	// }
	pkey, err := passkey.New(
		passkey.Config{
			WebauthnConfig: &webauthn.Config{
				RPDisplayName: "Misskey plugin repo",
				RPID:          "localhost",
				RPOrigins:     []string{"http://localhost:8080"},
			},
			UserStore:     store,
			SessionStore:  store,
			SessionMaxAge: time.Hour * 24,
		},
		passkey.WithLogger(logrus.StandardLogger()),
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to configure passkeys")
	}

	httpServer, err := server.NewServer(
		fswrapper.NewFSWrapper(frontendFS, "frontend/build/", false),
		store,
		pkey,
	)
	if err != nil {
		panic(err)
	}
	panic(httpServer.Run(":8080"))
}

func setLogLevelFromEnv() {
	level := os.Getenv("MK_REPO_LOG_LEVEL")
	fmt.Printf("Log level received from env: %s\n", level)
	switch level {
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
		logrus.SetLevel(logrus.InfoLevel)
	}
}
