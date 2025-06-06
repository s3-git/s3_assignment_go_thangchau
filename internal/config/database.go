package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	MigrationsPath  string        `json:"migrations_path"`
}

// loadDatabaseConfig loads database configuration from environment variables
func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "password"),
		Database:        getEnv("DB_NAME", "assignment-db"),
		SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "./migrations"),
	}
}

// Validate validates database configuration
func (d *DatabaseConfig) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if d.Port <= 0 || d.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	if d.User == "" {
		return fmt.Errorf("database user is required")
	}
	if d.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if d.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be greater than 0")
	}
	if d.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections must be non-negative")
	}
	if d.MaxIdleConns > d.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot be greater than max open connections")
	}
	return nil
}

// ConnectionString returns the PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode,
	)
}

// ConnectionStringWithoutPassword returns the connection string without password for logging
func (d *DatabaseConfig) ConnectionStringWithoutPassword() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Database, d.SSLMode,
	)
}