package config

import (
	"github.com/OpenTreeMap/otm-ecoservice/eco"
	"os"
)

type Config struct {
	Database   eco.DBInfo
	DataPath   string
	ServerHost string
	ServerPort string
}

func getEnvOrDefault(name string, defaultVal string) string {
	val := os.Getenv(name)
	if val != "" {
		return val
	}
	return defaultVal
}

func LoadConfig() Config {
	return Config{
		Database: eco.DBInfo{
			User:     getEnvOrDefault("OTM_DB_USER", "otm"),
			Password: getEnvOrDefault("OTM_DB_PASSWORD", "otm"),
			Database: getEnvOrDefault("OTM_DB_NAME", "otm"),
			Host:     getEnvOrDefault("OTM_DB_HOST", "localhost"),
		},
		DataPath:   getEnvOrDefault("OTM_ECO_DATA_DIR", "../data/"),
		ServerHost: getEnvOrDefault("OTM_ECO_HOST", "127.0.0.1"),
		ServerPort: getEnvOrDefault("OTM_ECO_PORT", "13000"),
	}
}

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
