package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DockerConfig holds Docker configuration
type DockerConfig struct {
	Host    string `mapstructure:"host"`
	Version string `mapstructure:"version"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DataDir string `mapstructure:"data_dir"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Docker: DockerConfig{
			Host:    "unix:///var/run/docker.sock",
			Version: "1.41",
		},
		Storage: StorageConfig{
			DataDir: "./data",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Set config file path
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Look for config in current directory and home directory
		viper.SetConfigName("orca")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.orca")
	}

	// Environment variables
	viper.SetEnvPrefix("ORCA")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config dosyası okunamadı: %w", err)
		}
		// Config file not found, use defaults
	}

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("config parse edilemedi: %w", err)
	}

	// Validate and create directories
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config doğrulanamadı: %w", err)
	}

	return config, nil
}

// validateConfig validates configuration and creates necessary directories
func validateConfig(config *Config) error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(config.Storage.DataDir, 0755); err != nil {
		return fmt.Errorf("data dizini oluşturulamadı: %w", err)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}

	if !validLevels[config.Logging.Level] {
		return fmt.Errorf("geçersiz log seviyesi: %s", config.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[config.Logging.Format] {
		return fmt.Errorf("geçersiz log formatı: %s", config.Logging.Format)
	}

	return nil
}

// GetConfigDir returns the configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".orca"), nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	viper.Set("server", config.Server)
	viper.Set("docker", config.Docker)
	viper.Set("storage", config.Storage)
	viper.Set("logging", config.Logging)

	if configPath == "" {
		configDir, err := GetConfigDir()
		if err != nil {
			return fmt.Errorf("config dizini alınamadı: %w", err)
		}
		
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("config dizini oluşturulamadı: %w", err)
		}
		
		configPath = filepath.Join(configDir, "orca.yaml")
	}

	return viper.WriteConfigAs(configPath)
}