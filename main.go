package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/logger"
	"github.com/go-pkgz/auth/token"
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

	// Setup authentication layer
	authOpts := authConfigFromGlobalConfig()
	authService := auth.NewService(authOpts)
	configureAuthServiceWithConfig(authService)

	logrus.WithField("mode", authMode).Infoln("Using authmode")
	switch authMode {
	case "prod":
		// Prod does nothing for now
	case "dev":
		// authService.AddProvider("dev", "", "")
		authService.AddDevProvider("localhost", 9090)
		// Also start dev server here
		// Dev server runs on port 8084
		go func() {
			devAuthServer, err := authService.DevAuth() // peak dev oauth2 server
			devAuthServer.Provider.URL = "localhost:8080"
			if err != nil {
				logrus.WithError(err).Fatalln("Failed to start oauth dev server")
			}
			devAuthServer.Run(context.Background())
		}()
	case "none":
		authService = nil
	}

	httpServer, err := server.NewServer(
		fswrapper.NewFSWrapper(frontendFS, "frontend/build/", false),
		&store,
		authService,
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

func authConfigFromGlobalConfig() auth.Opts {
	authOptions := auth.Opts{
		SecretReader: token.SecretFunc(func(aud string) (string, error) {
			return aud + "_super_secret", nil
		}),
		TokenDuration:     time.Minute,
		CookieDuration:    time.Hour * 24,
		Issuer:            "mk-plugin-repo",
		URL:               config.GlobalConfig.General.RootUrl,
		AvatarStore:       avatar.NewLocalFS("/tmp/avatars"),
		AvatarResizeLimit: 200,
		ClaimsUpd: token.ClaimsUpdFunc(func(claims token.Claims) token.Claims {
			if claims.User != nil && claims.User.Name == "dev_admin" {
				claims.User.SetAdmin(true)
			}
			return claims
		}),
		Validator: token.ValidatorFunc(
			func(_ string, claims token.Claims) bool { // rejects some tokens
				if claims.User != nil {
					if strings.HasPrefix(
						claims.User.ID,
						"github_",
					) { // allow all users with github auth
						return true
					}
					if strings.HasPrefix(
						claims.User.ID,
						"microsoft_",
					) { // allow all users with ms auth
						return true
					}
					if strings.HasPrefix(
						claims.User.ID,
						"patreon_",
					) { // allow all users with ms auth
						return true
					}
					if strings.HasPrefix(
						claims.User.Name,
						"dev_",
					) { // non-guthub allow only dev_* names
						return true
					}
					return strings.HasPrefix(claims.User.Name, "yeah no, this wont hit")
				}
				return false
			},
		),
		Logger:          logger.Func(logrus.Printf),
		UseGravatar:     true,
		SecureCookies:   authMode == "prod",
		DisableXSRF:     authMode != "prod",
		AvatarRoutePath: "/api/avatars",
	}
	return authOptions
}

func configureAuthServiceWithConfig(service *auth.Service) {
	if config.GlobalConfig.OAuthConfig == nil {
		return
	}
	if config.GlobalConfig.OAuthConfig.Github != nil {
		service.AddProvider(
			"github",
			config.GlobalConfig.OAuthConfig.Github.ID,
			config.GlobalConfig.OAuthConfig.Github.Secret,
		)
	}
	if config.GlobalConfig.OAuthConfig.Twitter != nil {
		service.AddProvider(
			"twitter",
			config.GlobalConfig.OAuthConfig.Twitter.ID,
			config.GlobalConfig.OAuthConfig.Twitter.Secret,
		)
	}
	if config.GlobalConfig.OAuthConfig.Microsoft != nil {
		service.AddProvider(
			"microsoft",
			config.GlobalConfig.OAuthConfig.Microsoft.ID,
			config.GlobalConfig.OAuthConfig.Microsoft.Secret,
		)
	}
	if config.GlobalConfig.OAuthConfig.Patreon != nil {
		service.AddProvider(
			"patreon",
			config.GlobalConfig.OAuthConfig.Patreon.ID,
			config.GlobalConfig.OAuthConfig.Patreon.Secret,
		)
	}
	if config.GlobalConfig.OAuthConfig.Google != nil {
		service.AddProvider(
			"google",
			config.GlobalConfig.OAuthConfig.Google.ID,
			config.GlobalConfig.OAuthConfig.Google.Secret,
		)
	}
}
