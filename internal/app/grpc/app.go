package grpcapp

import (
	"fmt"
	"log/slog"
	authgrpc "main/internal/grpc/auth"
	exchangegrpc "main/internal/grpc/exchange"
	walletgrpc "main/internal/grpc/wallet"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	auth authgrpc.Auth,
	wall walletgrpc.Wallet,
	exchange exchangegrpc.Exchange,
	port int,
) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.RegisterUser(gRPCServer, auth)
	walletgrpc.FinancialService(gRPCServer, wall)
	exchangegrpc.ExchangeWallet(gRPCServer, exchange)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.App.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is starting", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("grpc server is stopping", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
