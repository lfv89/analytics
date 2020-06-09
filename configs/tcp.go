package configs

import "os"

func GetPort(defaultPort string) string {
	value := os.Getenv("PORT")

	if len(value) == 0 {
		return defaultPort
	}

	return value
}
