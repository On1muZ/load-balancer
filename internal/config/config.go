package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Algorithm string

const (
	RoundRobin               Algorithm = "round_robin"
	WeightedRandom           Algorithm = "weighted_random"
	LeastConnections         Algorithm = "least_connections"
	WeightedLeastConnections Algorithm = "weighted_least_connections"
	P2C                      Algorithm = "p2c"
)

type App struct {
	Env string
}

type Server struct {
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

type HealthChecker struct {
	Timeout  time.Duration
	Interval time.Duration
}

type Config struct {
	App           *App
	Server        *Server
	HealthChecker *HealthChecker
	Algorithm     Algorithm
	NodesFilePath string
}

func Load() (*Config, error) {
	cfg := &Config{
		App:           &App{Env: getEnvString("APP_ENV", "dev")},
		NodesFilePath: getEnvString("NODES_FILE_PATH", ""),
		Server:        new(Server),
		HealthChecker: new(HealthChecker),
	}
	port, err := getEnvInt("HTTP_PORT", 9000)
	if err != nil {
		return nil, fmt.Errorf("error getting http port from env")
	}
	healthCheckerTimeout, err := getEnvInt("HEALTH_CHECKER_TIMEOUT", 1)
	if err != nil {
		return nil, fmt.Errorf("error parsing health checker timeout")
	}
	healthCheckerInterval, err := getEnvInt("HEALTH_CHECKER_INTERVAL", 1)
	if err != nil {
		return nil, fmt.Errorf("error parsing health checker interval")
	}
	serverReadHeaderTimeout, err := getEnvInt("SERVER_READ_HEADER_TIMEOUT", 5)
	if err != nil {
		return nil, fmt.Errorf("error parsing server read header timeout")
	}
	serverReadTimeout, err := getEnvInt("SERVER_READ_TIMEOUT", 30)
	if err != nil {
		return nil, fmt.Errorf("error parsingn server read timeout")
	}
	serverWriteTimeout, err := getEnvInt("SERVER_WRITE_TIMEOUT", 60)
	if err != nil {
		return nil, fmt.Errorf("error parsing server write timeout")
	}
	serverIdleTimeout, err := getEnvInt("SERVER_IDLE_TIMEOUT", 120)
	if err != nil {
		return nil, fmt.Errorf("error parsing server idle timeout")
	}
	serverMaxHeaderBytes, err := getEnvInt("MAX_HEADER_BYTES", 1048576)
	if err != nil {
		return nil, fmt.Errorf("error parsing server max header bytes")
	}
	var algorithm Algorithm
	switch getEnvString("ALGORITHM", "") {
	case "round_robin":
		algorithm = RoundRobin
	case "weighted_random":
		algorithm = WeightedRandom
	case "least_connections":
		algorithm = LeastConnections
	case "weighted_least_connections":
		algorithm = WeightedLeastConnections
	case "p2c":
		algorithm = P2C
	default:
		algorithm = ""
	}
	cfg.Algorithm = algorithm
	cfg.Server.Port = port
	cfg.Server.ReadHeaderTimeout = time.Duration(serverReadHeaderTimeout) * time.Second
	cfg.Server.ReadTimeout = time.Duration(serverReadTimeout) * time.Second
	cfg.Server.WriteTimeout = time.Duration(serverWriteTimeout) * time.Second
	cfg.Server.IdleTimeout = time.Duration(serverIdleTimeout) * time.Second
	cfg.Server.MaxHeaderBytes = serverMaxHeaderBytes
	cfg.HealthChecker.Interval = time.Duration(healthCheckerInterval) * time.Second
	cfg.HealthChecker.Timeout = time.Duration(healthCheckerTimeout) * time.Second
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Algorithm == "" {
		return fmt.Errorf("incorrect algorithm value")
	}
	if cfg.App.Env != "dev" && cfg.App.Env != "prod" {
		return fmt.Errorf("incorrect env value")
	}
	if cfg.NodesFilePath == "" {
		return fmt.Errorf("invalid nodes file path")
	}
	return nil
}

func getEnvString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	if v := os.Getenv(key); v != "" {
		return strconv.Atoi(v)
	}
	return fallback, nil
}
