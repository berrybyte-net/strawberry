package config

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CertDirectory string `toml:"cert_directory"`
	ACME          ACME   `toml:"acme"`
	Redis         Redis  `toml:"redis"`
}

type ACME struct {
	Email        string `toml:"email"`
	DirectoryURL string `toml:"directory_url"`
}

type Redis struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

func ParseFile(p string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(filepath.Clean(p), &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
