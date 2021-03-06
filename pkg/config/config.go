package config

import (
	"context"
	"time"
)

// Log defines the available logging configuration.
type Log struct {
	Level  string
	Pretty bool
	Color  bool
	File   string
}

// Debug defines the available debug configuration.
type Debug struct {
	Addr   string
	Token  string
	Pprof  bool
	Zpages bool
}

// HTTP defines the available http configuration.
type HTTP struct {
	Addr      string
	Namespace string
	Root      string
	CacheTTL  int
}

// GRPC defines the available grpc configuration.
type GRPC struct {
	Addr      string
	Namespace string
}

// Server configures a server.
type Server struct {
	Version string
	Name    string
}

// Tracing defines the available tracing configuration.
type Tracing struct {
	Enabled   bool
	Type      string
	Endpoint  string
	Collector string
	Service   string
}

// Asset defines the available asset configuration.
type Asset struct {
	Path string
}

// TokenManager is the config for using the reva token manager
type TokenManager struct {
	JWTSecret string
	TokenTTL  time.Duration
}

type WopiServer struct {
	Host        string
	Insecure    bool
	RevaGateway string
	IOPSecret   string
}

// Config combines all available configuration parts.
type Config struct {
	File         string
	Log          Log
	Debug        Debug
	HTTP         HTTP
	Server       Server
	Tracing      Tracing
	Asset        Asset
	TokenManager TokenManager

	WopiServer WopiServer

	Context    context.Context
	Supervised bool
}

// New initializes a new configuration with or without defaults.
func New() *Config {
	return &Config{}
}
