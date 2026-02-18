package main

import (
	"log/slog"
	"os"

	"admin.com/admin-api/config"
	"admin.com/admin-api/internal/app"
	dbpostgres "admin.com/admin-api/internal/repository/postgres"
	"admin.com/admin-api/pkg/logger"
)

func main() {
	logger.Init("info", "json")

	appCfg, err := config.Load()
	if err != nil {
		slog.Error(logger.MsgInvalidConfiguration, "error", err)
		os.Exit(1)
	}

	logger.Init(appCfg.LogLevel, appCfg.LogFormat)

	dbConn, err := dbpostgres.NewPostgresDB(appCfg.DatabaseDSN)
	if err != nil {
		slog.Error(logger.MsgDatabaseInitFailed, "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			slog.Error(logger.MsgDatabaseCloseFailed, "error", err)
		}
	}()

	httpHandler, err := app.NewHandler(appCfg, dbConn)
	if err != nil {
		slog.Error(logger.MsgServerFailed, "error", err)
		os.Exit(1)
	}
	httpServer := app.NewServer(appCfg, httpHandler)

	slog.Info(logger.MsgServerStarted, "address", appCfg.ServerAddress)

	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error(logger.MsgServerFailed, "error", err)
		os.Exit(1)
	}
}
