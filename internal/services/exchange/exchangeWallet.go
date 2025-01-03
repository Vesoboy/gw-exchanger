package exchange

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// ==================WALLET====================

func NewExchange(
	log *slog.Logger,
	exchCurrency ExchangeCurrency,
	exchRate GetExchangeRates,
	tokenTTL time.Duration,
) *Exchange {
	return &Exchange{
		log:          log,
		exchCurrency: exchCurrency,
		exchRate:     exchRate,
		tokenTTL:     tokenTTL,
	}
}

type Exchange struct {
	log          *slog.Logger
	exchCurrency ExchangeCurrency
	exchRate     GetExchangeRates
	tokenTTL     time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ExchangeCurrency interface {
	ExchangeCurrency(ctx context.Context,
		token string,
		from_currency string,
		to_currency string,
		amount float32,
	) (string,
		float32,
		map[string]float32,
		error)
}

type GetExchangeRates interface {
	GetExchangeRates(ctx context.Context,
		token string,
	) (string, map[string]float32, error)
}

func (e *Exchange) ExchangeCurrency(ctx context.Context, token string,
	from_currency string, to_currency string, amount float32) (string, float32, map[string]float32, error) {

	const op = "exchange.ExchangeCurrency"
	log := e.log.With(
		slog.String("op", op),
		slog.String("token", token),
		slog.String("from_currency", from_currency),
		slog.String("to_currency", to_currency),
		slog.String("amount", fmt.Sprintf("%f", amount)),
	)
	log.Info("Exchange currency")
	if e.exchCurrency == nil {
		return "", 0, nil, errors.New("ExchangeCurrency is not initialized")
	}

	message, exchAmount, balance, err := e.exchCurrency.ExchangeCurrency(ctx, token, from_currency, to_currency, amount)
	if err != nil {
		log.Error("failed to exchange wallet", err)
		return "", 0, nil, err
	}
	log.Info("Exchange OK")
	return message, exchAmount, balance, nil
}

func (e *Exchange) GetExchangeRates(ctx context.Context, token string) (string, map[string]float32, error) {

	const op = "exchange.GetExchangeRates"
	log := e.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)
	log.Info("Exchange currency")
	if e.exchRate == nil {
		return "", nil, errors.New("ExchangeRates is not initialized")
	}

	message, balance, err := e.exchRate.GetExchangeRates(ctx, token)
	if err != nil {

		log.Error("failed to get exchange rate", err)
		return "", nil, err
	}
	log.Info("Get Rate OK")
	return message, balance, nil
}
