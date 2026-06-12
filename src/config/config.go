package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds server configuration. All fields are CLI-only.
type Config struct {
	Port  int    // --port
	Host  string // --host
	Token string // --token (empty = no auth)
	Root  string // --root (required)
}

func Defaults() *Config {
	return &Config{
		Port: 7000,
		Host: "0.0.0.0",
	}
}

// ParseFlags parses CLI flags into cfg.
func ParseFlags(cfg *Config) {
	root := flag.String("root", "", "file serving root directory (required)")
	port := flag.Int("port", cfg.Port, "listen port")
	host := flag.String("host", cfg.Host, "listen address")
	token := flag.String("token", cfg.Token, "bearer token for authentication (empty disables auth)")

	flag.Parse()

	if *root == "" {
		fmt.Fprintln(os.Stderr, "error: --root is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg.Root = *root
	cfg.Port = *port
	cfg.Host = *host
	cfg.Token = *token
}
