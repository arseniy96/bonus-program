package main

import (
	"github.com/arseniy96/bonus-program/internal/config"
	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/router"
	"github.com/arseniy96/bonus-program/internal/server"
	"github.com/arseniy96/bonus-program/internal/store"
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

	store, err := store.NewStore(settings.DatabaseURI)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	s := server.NewServer(store, settings)
	r := router.NewRouter(s)

	logger.Log.Infow("start server", "host", settings.Host)
	return r.Run(settings.Host)
}
