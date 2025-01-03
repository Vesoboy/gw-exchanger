package auth

import (
	"context"
	"errors"
	"main/internal/storage"

	user "github.com/Vesoboy/proto-exchange/v2/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	LoginUser(ctx context.Context,
		email string,
		password string,
	) (token string, err error)

	RegisterUser(ctx context.Context,
		username string,
		email string,
		password string,
	) (message string, err error)
}

type serverAPI struct {
	user.UnimplementedAuthServer
	auth Auth
}

func RegisterUser(gRPC *grpc.Server, auth Auth) {
	user.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) LoginUser(
	ctx context.Context,
	req *user.LoginRequest,
) (*user.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}

	token, err := s.auth.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) RegisterUser(
	ctx context.Context,
	req *user.RegisterRequest,
) (*user.RegisterResponse, error) {

	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is empty")
	}
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}

	uid, err := s.auth.RegisterUser(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.RegisterResponse{
		Message: uid,
	}, nil
}
