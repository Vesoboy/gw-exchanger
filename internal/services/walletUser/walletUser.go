package walletuser

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/storage"
	"time"
)

// ==================WALLET====================

func NewWallet(
	log *slog.Logger,
	getBalanc GetBalance,
	deposit Deposit,
	withdraw Withdraw,
	tokenTTL time.Duration,
) *Wallet {
	return &Wallet{
		log:       log,
		getBalanc: getBalanc,
		deposit:   deposit,
		withdraw:  withdraw,
		tokenTTL:  tokenTTL,
	}
}

type Wallet struct {
	log       *slog.Logger
	getBalanc GetBalance
	deposit   Deposit
	withdraw  Withdraw
	tokenTTL  time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type GetBalance interface {
	GetBalance(
		ctx context.Context,
		token string,
	) (map[string]float32, error)
}

type Deposit interface {
	Deposit(
		ctx context.Context,
		token string,
		amount float32,
		currency string,
	) (string, map[string]float32, error)
}

type Withdraw interface {
	Withdraw(
		ctx context.Context,
		token string,
		amount float32,
		currency string,
	) (string, map[string]float32, error)
}

func (w *Wallet) GetBalance(ctx context.Context, token string) (map[string]float32, error) {

	const op = "walletUser.GetBalance"
	log := w.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)
	log.Info("Get balance")
	if w.getBalanc == nil {
		return nil, errors.New("GetBalance is not initialized")
	}

	balance, err := w.getBalanc.GetBalance(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("User already exists", err)
			return nil, fmt.Errorf("%w", ErrInvalidCredentials)
		}
		log.Error("failed to save user", err)
		return nil, err
	}
	log.Info("Кошелек найден")
	return balance, nil
}

func (w *Wallet) Deposit(ctx context.Context, token string, amount float32, currency string) (string, map[string]float32, error) {

	const op = "walletUser.Deposit"
	log := w.log.With(
		slog.String("op", op),
		slog.String("token", token),
		slog.String("amount", fmt.Sprint(amount)),
		slog.String("currency", currency),
	)
	log.Info("Deposit")
	if w.deposit == nil {
		return "", nil, errors.New("Deposit is not initialized")
	}

	message, balance, err := w.deposit.Deposit(ctx, token, amount, currency)
	if err != nil {

		log.Error("failed to deposit wallet", err)
		return "", nil, err
	}
	log.Info("Deposit OK")
	return message, balance, nil
}

func (w *Wallet) Withdraw(ctx context.Context, token string, amount float32, currency string) (string, map[string]float32, error) {

	const op = "walletUser.Withdraw"
	log := w.log.With(
		slog.String("op", op),
		slog.String("token", token),
		slog.String("amount", fmt.Sprint(amount)),
		slog.String("currency", currency),
	)
	log.Info("Withdraw")
	if w.withdraw == nil {
		return "", nil, errors.New("Withdraw is not initialized")
	}

	message, balance, err := w.withdraw.Withdraw(ctx, token, amount, currency)
	if err != nil {

		log.Error("failed to deposit wallet", err)
		return "", nil, err
	}
	log.Info("Deposit OK")
	return message, balance, nil
}
