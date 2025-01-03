package main

import (
	"fmt"
	"log/slog"
	"main/internal/app"
	"main/internal/config"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {

	cfg := config.MustLoad()

	logFile, log := setupLogger(cfg.Dev)
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	log.Info("starting wallet",
		slog.String("dev", cfg.Dev),
		slog.Any("cfg", cfg),
		slog.Int("port", cfg.GRPC.Port))

	application := app.New(log, cfg.GRPC.Port, cfg.Storage, cfg.Token)

	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("Application stopped", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()
	log.Info("Application stopped")
}

func setupLogger(env string) (*os.File, *slog.Logger) {
	var log *slog.Logger
	var logFile *os.File
	var err error

	if env == envDev {
		logFile, err = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Не удалось открыть файл для логов: %v\n", err)
			os.Exit(1)
		}
	} else {
		logFile = os.Stdout
	}

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return logFile, log
}
