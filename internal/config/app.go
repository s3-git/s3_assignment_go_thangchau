package config

// AppConfig holds general application configuration
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// loadAppConfig loads application configuration from environment variables
func loadAppConfig() AppConfig {
	return AppConfig{
		Name:        getEnv("APP_NAME", "assignment-api"),
		Version:     getEnv("APP_VERSION", "1.0.0"),
		Environment: getEnv("APP_ENV", "development"),
		Debug:       getEnvAsBool("APP_DEBUG", false),
	}
}

// IsDevelopment returns true if the environment is development
func (a *AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}

// IsProduction returns true if the environment is production
func (a *AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

// IsTest returns true if the environment is test
func (a *AppConfig) IsTest() bool {
	return a.Environment == "test"
}