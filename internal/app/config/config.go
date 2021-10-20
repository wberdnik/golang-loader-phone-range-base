package config

// Config  struct
type Config struct {
	LogLevel    string `toml:"log_level"`
	DatabaseDSN string `toml:"database_dsn"`
}

// Default config
func NewConfig() *Config {
	return &Config{
		LogLevel: "debug",
	}
}
