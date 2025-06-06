package config

import (
	"fmt"
	"strings"
)

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`    // megabytes
	MaxBackups int    `json:"max_backups"` // number of old log files to retain
	MaxAge     int    `json:"max_age"`     // days
	Compress   bool   `json:"compress"`
}

// loadLoggingConfig loads logging configuration from environment variables
func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:      getEnv("LOG_LEVEL", "info"),
		Format:     getEnv("LOG_FORMAT", "json"),
		Output:     getEnv("LOG_OUTPUT", "stdout"),
		File:       getEnv("LOG_FILE", "logs/app.log"),
		MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
		MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
		MaxAge:     getEnvAsInt("LOG_MAX_AGE", 28),
		Compress:   getEnvAsBool("LOG_COMPRESS", true),
	}
}

// Validate validates logging configuration
func (l *LoggingConfig) Validate() error {
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	level := strings.ToLower(l.Level)
	isValidLevel := false
	for _, validLevel := range validLevels {
		if level == validLevel {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		return fmt.Errorf("invalid log level: %s, must be one of: %s", l.Level, strings.Join(validLevels, ", "))
	}

	validFormats := []string{"json", "text"}
	format := strings.ToLower(l.Format)
	isValidFormat := false
	for _, validFormat := range validFormats {
		if format == validFormat {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		return fmt.Errorf("invalid log format: %s, must be one of: %s", l.Format, strings.Join(validFormats, ", "))
	}

	validOutputs := []string{"stdout", "stderr", "file"}
	output := strings.ToLower(l.Output)
	isValidOutput := false
	for _, validOutput := range validOutputs {
		if output == validOutput {
			isValidOutput = true
			break
		}
	}
	if !isValidOutput {
		return fmt.Errorf("invalid log output: %s, must be one of: %s", l.Output, strings.Join(validOutputs, ", "))
	}

	if l.Output == "file" && l.File == "" {
		return fmt.Errorf("log file path is required when output is set to file")
	}

	if l.MaxSize <= 0 {
		return fmt.Errorf("log max size must be positive")
	}

	if l.MaxBackups < 0 {
		return fmt.Errorf("log max backups must be non-negative")
	}

	if l.MaxAge < 0 {
		return fmt.Errorf("log max age must be non-negative")
	}

	return nil
}

// IsDebugLevel returns true if the log level is debug
func (l *LoggingConfig) IsDebugLevel() bool {
	return strings.ToLower(l.Level) == "debug"
}

// IsJSONFormat returns true if the log format is JSON
func (l *LoggingConfig) IsJSONFormat() bool {
	return strings.ToLower(l.Format) == "json"
}

// IsFileOutput returns true if the log output is file
func (l *LoggingConfig) IsFileOutput() bool {
	return strings.ToLower(l.Output) == "file"
}