package exchange

import (
	"context"
	"errors"
	"main/internal/storage"

	"github.com/Vesoboy/proto-exchange/v2/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Exchange interface {
	//обмен валюты
	ExchangeCurrency(ctx context.Context,
		token string,
		from_currency string,
		to_currency string,
		amount float32,
	) (string,
		float32,
		map[string]float32,
		error)

	// Получение курсов обмена всех валют
	GetExchangeRates(ctx context.Context,
		token string,
	) (string, map[string]float32, error)
}

type exchangeAPI struct {
	user.UnimplementedExchangeServiceServer
	exchange Exchange
}

func ExchangeWallet(gRPC *grpc.Server, exchange Exchange) {
	user.RegisterExchangeServiceServer(gRPC, &exchangeAPI{exchange: exchange})
}

func (e *exchangeAPI) GetExchangeRates(
	ctx context.Context,
	req *user.RatesRequest,
) (*user.ExchangeRatesResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Token is empty")
	}

	message, rates, err := e.exchange.GetExchangeRates(ctx, req.GetToken())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return &user.ExchangeRatesResponse{
		Message: message,
		Rates:   rates,
	}, nil
}

func (e *exchangeAPI) ExchangeCurrency(
	ctx context.Context,
	req *user.ExchangeRequest,
) (*user.TransactionResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Token is empty")
	}
	if req.GetFromCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "FromCurrency is empty")
	}
	if req.GetToCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "ToCurrency is empty")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Amount is empty")
	}

	message, amount, balance, err := e.exchange.ExchangeCurrency(ctx, req.GetToken(), req.GetFromCurrency(), req.GetToCurrency(), req.GetAmount())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return &user.TransactionResponse{
		Message:       message,
		AmountFromTo:  amount,
		BalanceFromTo: balance,
	}, nil

}
