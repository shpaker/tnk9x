package config

// Config содержит конфигурацию приложения
type Config struct {
	AppConfig
	LogConfig
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		AppConfig{
			Name:         getEnv("APP_NAME", "gonflict"),
			ScreenWidth:  640,
			ScreenHeight: 480,
		},
		LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	return cfg, nil
}
