package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all server configuration.
// Root is NOT in the config file — it must be specified via --root CLI flag.
type Config struct {
	Port       int    `json:"port"`
	Host       string `json:"host"`
	Token      string `json:"token"`
	Theme      string `json:"theme"`
	FontFamily string `json:"font_family"`
	FontSize   int    `json:"font_size"`

	Root       string // CLI only, not in JSON
	configPath string
}

func Defaults() *Config {
	return &Config{
		Root:       "./",
		Port:       7000,
		Host:       "0.0.0.0",
		Token:      "",
		Theme:      "light",
		FontFamily: "Ubuntu Sans Mono",
		FontSize:   14,
	}
}

func (c *Config) ConfigPath() string { return c.configPath }

// Load loads config from a JSON file. Returns defaults if file does not exist.
func Load(path string) (*Config, error) {
	cfg := Defaults()
	absPath, err := filepath.Abs(path)
	if err != nil {
		return cfg, fmt.Errorf("config path: %w", err)
	}
	cfg.configPath = absPath

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (c *Config) Save() error {
	// Ensure config directory exists.
	if err := os.MkdirAll(filepath.Dir(c.configPath), 0o755); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	saveCopy := struct {
		Port       int    `json:"port"`
		Host       string `json:"host"`
		Token      string `json:"token"`
		Theme      string `json:"theme"`
		FontFamily string `json:"font_family"`
		FontSize   int    `json:"font_size"`
	}{
		Port:       c.Port,
		Host:       c.Host,
		Token:      c.Token,
		Theme:      c.Theme,
		FontFamily: c.FontFamily,
		FontSize:   c.FontSize,
	}
	data, err := json.MarshalIndent(saveCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(c.configPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// DefaultConfigPath returns ~/.config/remdev/remdev.json.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "rdev.json"
	}
	return filepath.Join(home, ".config", "rdev", "rdev.json")
}

// ParseFlagsAndMerge extracts the --config flag from os.Args, loads the config
// file, then parses all CLI flags and merges them into the config.
// CLI flags (--root, --port, --host) > --option > config file > defaults.
func ParseFlagsAndMerge(cfg *Config) []string {
	// Extract --config early to know which file to load.
	configPath := extractArg("config", DefaultConfigPath())
	loaded, err := Load(configPath)
	if err == nil {
		*cfg = *loaded
	}
	cfg.configPath = configPath

	// Now parse all flags using the config values as defaults.
	configFlag := flag.String("config", configPath, "path to config file")
	root := flag.String("root", cfg.Root, "file serving root directory (required)")
	port := flag.Int("port", cfg.Port, "listen port")
	host := flag.String("host", cfg.Host, "listen address")
	var options optionsFlag
	flag.Var(&options, "option", "override config key=value (repeatable)")

	flag.Parse()

	cfg.configPath = *configFlag

	// Root is always from CLI and not from config file.
	cfg.Root = *root

	// Apply explicit CLI flags.
	if isFlagSet("port") {
		cfg.Port = *port
	}
	if isFlagSet("host") {
		cfg.Host = *host
	}

	// Apply --option overrides.
	for _, opt := range options {
		cfg.applyOption(opt)
	}

	return flag.Args()
}

func extractArg(name, defaultVal string) string {
	prefix := "--" + name + "="
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	prefix2 := "--" + name
	for i, arg := range os.Args {
		if arg == prefix2 && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return defaultVal
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
	case "theme":
		c.Theme = val
	case "font_family":
		c.FontFamily = val
	case "font_size":
		fmt.Sscanf(val, "%d", &c.FontSize)
	}
}

type optionsFlag []string

func (o *optionsFlag) String() string { return strings.Join(*o, ", ") }
func (o *optionsFlag) Set(v string) error {
	*o = append(*o, v)
	return nil
}
