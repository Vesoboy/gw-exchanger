package wallet

import (
	"context"
	"errors"
	"main/internal/storage"

	"github.com/Vesoboy/proto-exchange/v2/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Wallet interface {
	GetBalance(ctx context.Context, token string) (map[string]float32, error)
	Deposit(ctx context.Context, token string, amount float32, currency string) (string, map[string]float32, error)
	Withdraw(ctx context.Context, token string, amount float32, currency string) (string, map[string]float32, error)
}

type walletAPI struct {
	user.UnimplementedFinancialServiceServer
	wallet Wallet
}

func FinancialService(gRPC *grpc.Server, wallet Wallet) {
	user.RegisterFinancialServiceServer(gRPC, &walletAPI{wallet: wallet})
}

func (w *walletAPI) GetBalance(
	ctx context.Context,
	req *user.GetBalanceRequest,
) (*user.BalanceResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Token is empty")
	}

	balance, err := w.wallet.GetBalance(ctx, req.GetToken())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	responseBalance := make(map[string]float32)
	for k, v := range balance {
		responseBalance[k] = v
	}

	return &user.BalanceResponse{
		Balance: balance,
	}, nil
}

func (w *walletAPI) Deposit(
	ctx context.Context,
	req *user.DepositRequest,
) (*user.WithdrawDepositResponse, error) {

	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "username is empty")
	}
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is empty")
	}
	if req.GetCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is empty")
	}

	message, depositBalance, err := w.wallet.Deposit(ctx, req.GetToken(), req.GetAmount(), req.GetCurrency())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &user.WithdrawDepositResponse{
		Message:    message,
		NewBalance: depositBalance,
	}, nil

}

func (w *walletAPI) Withdraw(
	ctx context.Context,
	req *user.WithdrawRequest,
) (*user.WithdrawDepositResponse, error) {

	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "username is empty")
	}
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is empty")
	}
	if req.GetCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is empty")
	}

	message, depositBalance, err := w.wallet.Withdraw(ctx, req.GetToken(), req.GetAmount(), req.GetCurrency())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &user.WithdrawDepositResponse{
		Message:    message,
		NewBalance: depositBalance,
	}, nil
}
