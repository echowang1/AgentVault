package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerHost string
	ServerPort int
	APIKeys    map[string]bool
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvInt("SERVER_PORT", 8080),
		APIKeys:    map[string]bool{},
	}

	apiKeyStr := getEnv("MPC_API_KEYS", "")
	if apiKeyStr != "" {
		for _, key := range split(apiKeyStr, ",") {
			k := trim(key)
			if k != "" {
				cfg.APIKeys[k] = true
			}
		}
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func split(s, sep string) []string {
	return strings.Split(s, sep)
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
