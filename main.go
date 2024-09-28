package main

import (
	"embed"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/mstarongithub/passkey"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
	setLogger()
	setLogLevelFromEnv()
	_, err := config.ReadConfig(nil)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Err reading config, using default and attempting to write it to default location")
		config.SetGlobalToDefault()
		config.WriteDefaultConfigToDefaultLocation()
	}

	err = util.CreateFileIfNotExists(DB_DEFAULT_FILE)
	if err != nil {
		panic(err)
	}

	store, err := storage.NewStorage(DB_DEFAULT_FILE, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup storage layer")
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
		passkey.WithLogger(&util.ZerologWrapper{}),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to configure passkeys")
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
	switch strings.ToLower(*flagLogLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func setLogger(extraLogWriters ...io.Writer) {
	if *flagPrettyPrint {
		console := zerolog.ConsoleWriter{Out: os.Stderr}
		log.Logger = zerolog.New(zerolog.MultiLevelWriter(append([]io.Writer{console}, extraLogWriters...)...)).
			With().
			Timestamp().
			Logger()
	} else {
		log.Logger = zerolog.New(zerolog.MultiLevelWriter(
			append([]io.Writer{log.Logger}, extraLogWriters...)...,
		)).With().Timestamp().Logger()
	}
}
