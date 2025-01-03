package storage

import "errors"

var (
	ErrUserNotFound       = errors.New("Пользователь не найден")
	ErrUserExists         = errors.New("Пользователь уже существует")
	ErrInvalidCredentials = errors.New("Неверные учетные данные")
)
