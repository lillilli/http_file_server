package config

// Config - service configuration
type Config struct {
	HTTP HTTPServer
	Log  LogConfig

	StaticDir string
}

// LogConfig - logger configuration
type LogConfig struct {
	MinLevel string `env:"LOG_LEVEL" default:"DEBUG"`
}

// HTTPServer - http server configuration
type HTTPServer struct {
	Host string `default:"0.0.0.0"`
	Port int    `default:"8081"`
}
