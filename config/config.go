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
	RootUrl       string `toml:"root_url"`
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
	Enabled  bool   `toml:"enabled"`
	Username string `toml:"username"`
	// Password is argon2id encrypted using the default settings of the github.com/ermites-io/passwd Argon2IdDefault config
	// The params this uses for Argon2Id are:
	// - Time:    1
	// - Memory:  64 * 1024
	// - Threads: 8
	// - SaltLen: 16
	// - Keylen:  32
	Password string `toml:"password"`
	// If set and true, the previous comment can be ignored as the password will be handled as if it was raw
	// Note: Pleas don't use this and instead offer an already hashed password for increased safety
	PasswordIsRaw *bool `toml:"password_is_raw"`
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
