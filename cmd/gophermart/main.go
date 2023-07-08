package main

import (
	"github.com/arseniy96/bonus-program/internal/config"
	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/router"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	settings := config.Initialize()

	if err := logger.Initialize(settings.LoggingLevel); err != nil {
		return err
	}

	// init database
	// create server

	r := router.NewRouter()
	logger.Log.Infow("start server", "host", settings.Host)

	return r.Run(settings.Host)
}
