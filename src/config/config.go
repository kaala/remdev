package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds server configuration. All fields are CLI-only.
type Config struct {
	Port  int    // --port
	Host  string // --host
	Token string // --option token=... (empty = no auth)
	Root  string // --root (required)
}

func Defaults() *Config {
	return &Config{
		Port: 7000,
		Host: "0.0.0.0",
	}
}

// ParseFlags parses CLI flags and returns remaining args.
func ParseFlags(cfg *Config) []string {
	root := flag.String("root", "", "file serving root directory (required)")
	port := flag.Int("port", cfg.Port, "listen port")
	host := flag.String("host", cfg.Host, "listen address")
	var options optionsFlag
	flag.Var(&options, "option", "override config key=value (repeatable)")

	flag.Parse()

	if *root == "" {
		fmt.Fprintln(os.Stderr, "error: --root is required")
		flag.Usage()
		os.Exit(1)
	}
	cfg.Root = *root

	if isFlagSet("port") {
		cfg.Port = *port
	}
	if isFlagSet("host") {
		cfg.Host = *host
	}

	for _, opt := range options {
		cfg.applyOption(opt)
	}

	return flag.Args()
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func (c *Config) applyOption(opt string) {
	parts := strings.SplitN(opt, "=", 2)
	if len(parts) != 2 {
		return
	}
	key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	switch key {
	case "port":
		fmt.Sscanf(val, "%d", &c.Port)
	case "host":
		c.Host = val
	case "token":
		c.Token = val
	}
}

type optionsFlag []string

func (o *optionsFlag) String() string { return strings.Join(*o, ", ") }
func (o *optionsFlag) Set(v string) error {
	*o = append(*o, v)
	return nil
}
