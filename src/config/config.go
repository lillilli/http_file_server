package config

// Config - определение структуры конфига
type Config struct {
	HTTP HTTPServer
	Log  LogConfig

	StaticDir string
}

// LogConfig конфигурация логгера
type LogConfig struct {
	Writer   string
	MinLevel string `env:"LOG_LEVEL"`
}

// HTTPServer - параметры HTTP-сервера
type HTTPServer struct {
	Host string `default:"0.0.0.0"`
	Port int    `default:"8081"`
}
