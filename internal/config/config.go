package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Settings struct {
	Host         string `env:"RUN_ADDRESS"`
	DatabaseURI  string `env:"DATABASE_URI"`
	AccrualHost  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LoggingLevel string `env:"LOG_LEVEL"`
}

func Initialize() *Settings {
	settings := &Settings{}

	flag.StringVar(&settings.Host, "a", "localhost:8081", "server host with port")
	flag.StringVar(&settings.DatabaseURI, "d", "", "database connection data")
	flag.StringVar(&settings.AccrualHost, "r", "localhost:8080", "accrual host")
	flag.StringVar(&settings.LoggingLevel, "l", "info", "log level")
	flag.Parse()

	err := env.Parse(settings)
	if err != nil {
		panic(err)
	}

	return settings
}
