package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

const DEFAULT_CONFIG_FILE_NAME = "config.toml"

var ErrNoConfig = fmt.Errorf("no config set")

type ConfigGeneral struct {
	// The protocol which is used to access the server publicly
	Protocol string
	// The top domain the server resides under publicly ("example.com")
	TopDomain string
	// The subdomain the server resides under publicy ("something".example.com)
	SubDomain *string
	// The port the server runs under, not necessarly the public one
	PrivatePort int
	// The port the server is accessible from in public, usually important for reverse proxies
	PublicPort    int
	HashingSecret string `toml:"hash_secret"`
}

type ConfigSSL struct {
	// Whether the server should handle ssl verification itself
	// If it's behind a router like nginx or traeffik, you probably want to disable this
	HandleSslInApp bool `toml:"handle_ssl_in_app"`
	// Whether to use LetsEncrypt for Ssl certificates. Only taken into account if HandleSslInApp is true
	UseLetsEncrypt *bool `toml:"use_lets_encrypt"`
	// The path to a custom certificate if UseLetsEncrypt is false
	// It is the certificate owner's responsibility to keep the certificate up to date
	CustomCertificatePath *string `toml:"custom_certificate_path"`
}

type ConfigWebauth struct {
	DisplayName string `toml:"display_name"`
	// ID can be extracted from root url
}

// Superuser data
// Note: Will be overwritten in dev mode since su and dev share the same ID of 0
type ConfigSuperuser struct {
	FirstSetupOTP string `toml:"first_setup_otp"`
}

type Config struct {
	// General config stuff. Required
	General ConfigGeneral `toml:"general"`
	// SSL Config. Required
	SslConfig ConfigSSL `toml:"ssl"`
	// Webauth config. Required
	WebAuth ConfigWebauth `toml:"webauth"`
	// Superuser account config. Required
	Superuser ConfigSuperuser `toml:"superuser"`
}

var defaultConfig = Config{
	General: ConfigGeneral{
		Protocol:      "http",
		TopDomain:     "localhost",
		SubDomain:     nil,
		PrivatePort:   8080,
		PublicPort:    8080,
		HashingSecret: "placeholder",
	},
	SslConfig: ConfigSSL{
		HandleSslInApp:        false,
		UseLetsEncrypt:        nil,
		CustomCertificatePath: nil,
	},
	WebAuth: ConfigWebauth{
		DisplayName: "Misskey Plugin Repo",
	},
	Superuser: ConfigSuperuser{
		FirstSetupOTP: "placeholder",
	},
}

var GlobalConfig Config

func ReadConfig(fileName *string) (Config, error) {
	if fileName == nil {
		return ReadFromFileName(DEFAULT_CONFIG_FILE_NAME, true)
	} else {
		return ReadFromFileName(*fileName, true)
	}
}

func ReadFromFileName(fileName string, writeToGlobal bool) (config Config, err error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return config, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}
	err = toml.Unmarshal(content, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse file %s as toml config: %w", fileName, err)
	}
	if writeToGlobal {
		GlobalConfig = config
	}
	return config, nil
}

func SetGlobalToDefault() {
	GlobalConfig = defaultConfig
}

func WriteDefaultConfigToDefaultLocation() {
	f, err := os.Create(DEFAULT_CONFIG_FILE_NAME)
	defer f.Close()
	if err != nil {
		log.Error().Err(err).Msg("Can't create default config file! Exiting")
		os.Exit(1)
	}
	err = toml.NewEncoder(f).Encode(&defaultConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write default config to default file! Exiting")
		os.Exit(1)
	}
}
