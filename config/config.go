package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const DEFAULT_CONFIG_FILE_NAME = "config.toml"

var ErrNoConfig = fmt.Errorf("no config set")

var GlobalConfig *Config

type ConfigGeneral struct {
	// The root domain where the entire project resides under
	// This includes port and protocol. Examples:
	// - http://localhost:8080
	// - https://subdomain.example.com
	RootUrl string `toml:"root_url"`
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

type ConfigAuthProvider struct {
	ID     string `toml:"id"`
	Secret string `toml:"secret"`
}

type ConfigOauth struct {
	Github    *ConfigAuthProvider `toml:"github,omitempty"`
	Twitter   *ConfigAuthProvider `toml:"twitter,omitempty"`
	Microsoft *ConfigAuthProvider `toml:"microsoft,omitempty"`
	Patreon   *ConfigAuthProvider `toml:"patreon,omitempty"`
	Google    *ConfigAuthProvider `toml:"google,omitempty"`
}

type Config struct {
	General ConfigGeneral `toml:"general"`
	// SSL Config. Required
	SslConfig ConfigSSL `toml:"ssl"`
	// OAuth config. Optional
	OAuthConfig *ConfigOauth `toml:"oauth"`
}

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
		GlobalConfig = &config
	}
	return config, nil
}

func SetGlobalToDefault() {
	GlobalConfig = &Config{
		SslConfig: ConfigSSL{
			HandleSslInApp: false,
		},
		General: ConfigGeneral{
			RootUrl: "localhost:8080",
		},
	}
}
