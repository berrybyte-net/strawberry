package config

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CertDirectory  string `toml:"cert_directory"`
	MaxBodyBytes   int    `toml:"max_body_bytes"`
	MaxHeaderBytes int    `toml:"max_header_bytes"`
	API            API    `toml:"api"`
	ACME           ACME   `toml:"acme"`
	Redis          Redis  `toml:"redis"`
}

type API struct {
	UseSSL     bool     `toml:"use_ssl"`
	Port       int      `toml:"port"`
	Token      string   `toml:"token"`
	AllowedIPs []string `toml:"allowed_ips"`
}

type ACME struct {
	Email        string `toml:"email"`
	DirectoryURL string `toml:"directory_url"`
}

type Redis struct {
	Prefix   string `toml:"prefix"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
}

// ParseFile returns a new Config that's decoded using the given path.
func ParseFile(p string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(filepath.Clean(p), &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
