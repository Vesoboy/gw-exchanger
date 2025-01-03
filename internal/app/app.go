package app

import (
	"log/slog"
	grpcapp "main/internal/app/grpc"

	"main/internal/services/auth"
	exchangewall "main/internal/services/exchange"
	walletuser "main/internal/services/walletUser"
	"main/internal/storage/postgresql"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := postgresql.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, tokenTTL)
	walService := walletuser.NewWallet(log, storage, storage, storage, tokenTTL)
	exchService := exchangewall.NewExchange(log, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, walService, exchService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
