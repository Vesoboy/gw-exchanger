package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/domain/models"
	"main/internal/lib/jwt"
	"main/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//=================AUTH====================

type serverAuth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		username string,
		email string,
		passHash []byte,
	) (message string, err error)
}

type UserProvider interface {
	User(
		ctx context.Context,
		email string,
	) (user models.User, err error)
}

// New returns new auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	tokenTTL time.Duration,
) *serverAuth {
	return &serverAuth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		tokenTTL:     tokenTTL,
	}
}

// LoginUser user
func (a *serverAuth) LoginUser(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "auth.LoginUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("Login user")
	if a.userProvider == nil {
		return "", errors.New("userProvider is not initialized")
	}
	user, err := a.userProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("User not found", err)
			return "", fmt.Errorf("%s: %w", op, storage.ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("Invalid password", err)
		return "", fmt.Errorf("%s: %w", op, storage.ErrInvalidCredentials)
	}

	token, err := jwt.NewToken(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

// RegisterUser New user
func (a *serverAuth) RegisterUser(
	ctx context.Context,
	username string,
	email string,
	password string,
) (string, error) {
	const op = "auth.RegisterUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("Registering new user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)
		return "", err
	}
	fmt.Print(passHash)

	uid, err := a.userSaver.SaveUser(ctx, username, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("User already exists", err)
			return "", fmt.Errorf("%s: %w", op, storage.ErrInvalidCredentials)
		}
		log.Error("failed to save user", err)
		return "", err
	}
	log.Info("User saved")

	return uid, nil
}
