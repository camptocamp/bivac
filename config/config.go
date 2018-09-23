package config

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

// Config stores the handler's configuration and UI interface parameters
type Config struct {
	Version    bool   `long:"version" description:"Display version."`
	LogLevel   string `short:"l" long:"log-level" description:"Set log level ('debug', 'info', 'warn', 'error', 'fatal', 'panic')." env:"BIVAC_LOG_LEVEL" default:"info"`
	JSON       bool   `short:"j" long:"json" description:"Log as JSON (to stderr)." env:"BIVAC_JSON_OUTPUT"`
	ServerMode bool   `short:"s" long:"server" description:"Use Bivac as server." env:"BIVAC_SERVER_MODE"`
	CLIMode    bool   `short:"c" logn:"cli" description:"Use Bivac CLI." env:"BIVAC_CLI_MODE"`
}

// LoadConfig loads the config from flags and environment
func LoadConfig(version string) *Config {
	var c Config
	parser := flags.NewParser(&c, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if c.Version {
		fmt.Printf("Bivac %v\n", version)
		os.Exit(0)
	}

	return &c
}
