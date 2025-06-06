package config

import (
	"fmt"
	"time"
)

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	GracefulStop time.Duration `json:"graceful_stop"`
	TLS          TLSConfig     `json:"tls"`
	CORS         CORSConfig    `json:"cors"`
	RateLimit    RateLimitConfig `json:"rate_limit"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
	AllowedCredentials bool `json:"allowed_credentials"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled    bool          `json:"enabled"`
	RequestsPerMinute int    `json:"requests_per_minute"`
	BurstSize  int           `json:"burst_size"`
}

// loadServerConfig loads server configuration from environment variables
func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host:         getEnv("SERVER_HOST", "0.0.0.0"),
		Port:         getEnvAsInt("SERVER_PORT", 8080),
		ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		GracefulStop: getEnvAsDuration("SERVER_GRACEFUL_STOP", 30*time.Second),
		TLS:          loadTLSConfig(),
		CORS:         loadCORSConfig(),
		RateLimit:    loadRateLimitConfig(),
	}
}

// loadTLSConfig loads TLS configuration
func loadTLSConfig() TLSConfig {
	return TLSConfig{
		Enabled:  getEnvAsBool("TLS_ENABLED", false),
		CertFile: getEnv("TLS_CERT_FILE", ""),
		KeyFile:  getEnv("TLS_KEY_FILE", ""),
	}
}

// loadCORSConfig loads CORS configuration
func loadCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:            getEnvAsBool("CORS_ENABLED", true),
		AllowedOrigins:     getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}, ","),
		AllowedMethods:     getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, ","),
		AllowedHeaders:     getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}, ","),
		AllowedCredentials: getEnvAsBool("CORS_ALLOWED_CREDENTIALS", false),
	}
}

// loadRateLimitConfig loads rate limiting configuration
func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:           getEnvAsBool("RATE_LIMIT_ENABLED", false),
		RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
		BurstSize:         getEnvAsInt("RATE_LIMIT_BURST_SIZE", 10),
	}
}

// Validate validates server configuration
func (s *ServerConfig) Validate() error {
	if s.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if s.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	if s.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	if s.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}
	if s.GracefulStop <= 0 {
		return fmt.Errorf("graceful stop timeout must be positive")
	}
	
	// Validate TLS
	if s.TLS.Enabled {
		if s.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file is required when TLS is enabled")
		}
		if s.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}
	
	// Validate rate limiting
	if s.RateLimit.Enabled {
		if s.RateLimit.RequestsPerMinute <= 0 {
			return fmt.Errorf("requests per minute must be positive when rate limiting is enabled")
		}
		if s.RateLimit.BurstSize <= 0 {
			return fmt.Errorf("burst size must be positive when rate limiting is enabled")
		}
	}
	
	return nil
}

// Address returns the server address in host:port format
func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}