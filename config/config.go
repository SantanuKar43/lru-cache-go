package config

import (
	"os"
	"strconv"
	"log"
	"github.com/SantanuKar43/lru-cache-go/cache"
)

type Config struct {
	Port int
	Capacity int
	WalSize int64
	FlushStrategy cache.FlushStrategy
}

func GetConfig() *Config {
	config := new(Config)
	var err error
	config.Port, err = strconv.Atoi(getEnvVar("LERU_PORT_NUMBER", "9090"))
	if err != nil || config.Port < 0 {
		log.Fatal("Error occurred while starting server, invalid port", err)
	}

	config.Capacity, err = strconv.Atoi(getEnvVar("LERU_CAPACITY", "3"))
	if err != nil || config.Capacity > 1000000 || config.Capacity < 1 {
		log.Fatal("Error occurred while starting server, invalid N", err)
	}

	config.WalSize, err = strconv.ParseInt(getEnvVar("LERU_WAL_SIZE", "100"), 10, 64) // 100 bytes
	if err != nil || config.WalSize < 1 {
		log.Fatal("Error occurred while starting server, invalid WAL size", err)
	}

	config.FlushStrategy = cache.FlushStrategy(getEnvVar("LERU_FLUSH_STRATEGY", "ALWAYS"))
	if config.FlushStrategy == "" {
		log.Fatal("Invalid flush strategy")
	}
	return config
}

func getEnvVar(key string, defaultVal string) string {
	val, present := os.LookupEnv(key)
	if present {
		return val
	}
	return defaultVal
}