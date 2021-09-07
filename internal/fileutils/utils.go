package fileutils

import "os"

func GetEnvOrDefault(envVar string, defaultVal string) string {
	envVal := os.Getenv(envVar)
	if envVal == "" {
		return defaultVal
	}
	return envVal
}
