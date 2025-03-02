package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// ServiceConfig represents a base configuration for all services
type ServiceConfig struct {
	Server     ServerConfig        `mapstructure:"server"`
	HttpServer HttpServerConfig    `mapstructure:"http_server"`
	Logging    logger.LoggerConfig `mapstructure:"logging"`
	Database   DatabaseConfig      `mapstructure:"database"`
	NATS       NATSConfig          `mapstructure:"nats"`
	Storage    Storage             `mapstructure:"storage"`
	JWT        JWT                 `mapstructure:"jwt"`
	Upload     Upload              `mapstructure:"upload"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type HttpServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Path     string `mapstructure:"path"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type NATSConfig struct {
	Servers []string `mapstructure:"servers"`
	Cluster string   `mapstructure:"cluster"`
}

type Storage struct {
	Provider    string `mapstructure:"provider"`
	BasePath    string `mapstructure:"base_path"`
	MaxFileSize int64  `mapstructure:"max_file_size"`
}

type Upload struct {
	MaxFileSize int64  `mapstructure:"max_file_size"`
	GRPCAddress string `mapstructure:"grpc_address"`
}

type JWT struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
}

// Load configuration with environment and service-specific overrides
func Load(serviceName string, defaults *ServiceConfig) (*ServiceConfig, error) {
	v := viper.New()

	// Set configuration sources and precedence
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Search paths
	configPaths := []string{
		".",
		"./configs",
		"/etc/" + serviceName,
		"$HOME/." + serviceName,
	}
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// Environment variable prefix
	v.SetEnvPrefix(serviceName)
	v.AutomaticEnv()

	// Bind environment variables to config
	v.BindEnv("server.host")
	v.BindEnv("server.port")
	v.BindEnv("logging.level")

	// Set defaults if provided
	if defaults != nil {
		v.SetDefault("server", defaults.Server)
		v.SetDefault("logging", defaults.Logging)
		v.SetDefault("database", defaults.Database)
		v.SetDefault("nats", defaults.NATS)
	}

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			return defaults, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config ServiceConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}
