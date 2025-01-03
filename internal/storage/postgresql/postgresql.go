package postgresql

import (
	"context"
	"errors"
	"fmt"
	"log"
	"main/internal/domain/models"
	"main/internal/storage"

	jwt "main/internal/lib/jwt"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.New"

	db, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = AutoMigrate(db)
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	// Обновляем курсы валют при инициализации
	err = UpdateExchangeRates(db)
	if err != nil {
		log.Printf("Ошибка обновления курсов валют: %v", err)
	} else {
		log.Println("Курсы валют успешно обновлены.")
	}
	return &Storage{db: db}, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &UserWallet{}, &ExchangeRate{})
}

// SaveUser saves user to db.
func (s *Storage) SaveUser(ctx context.Context, username, email string, passHash []byte) (string, error) {

	tx := s.db.Begin()
	id := uuid.New()

	query := `INSERT INTO users (id, username, email, pass_hash) VALUES ($1, $2, $3, $4)`
	if err := s.db.Exec(query, id, username, email, passHash).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("Ошибка создания пользователя: %v", err)
	}

	err := s.AddWalletUser(ctx, id)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("Ошибка создания кошелька: %v", err)
	}

	tx.Commit()
	return "success", nil
}

func GetExchangeRates(db *gorm.DB) ([]models.ExchangeRate, error) {

	// Обновляем курсы валют при запросе курсов
	err := UpdateExchangeRates(db)
	if err != nil {
		log.Printf("Ошибка обновления курсов валют: %v", err)
	} else {
		log.Println("Курсы валют успешно обновлены.")
	}

	curModel := []models.ExchangeRate{}
	err = db.Find(&curModel).Error
	if err != nil {
		return nil, fmt.Errorf("Ошибка получения списка валют: %v", err)
	}

	return curModel, nil
}

func (s *Storage) AddWalletUser(ctx context.Context, idUser uuid.UUID) error {

	currExch, err := GetExchangeRates(s.db)
	if err != nil {
		return fmt.Errorf("Ошибка получения списка валют: %v", err)
	}

	fmt.Printf("Список валют: %v\n", currExch)

	iduuid := make([]uuid.UUID, len(currExch))
	for i := range currExch {
		iduuid[i] = uuid.New()
	}

	query := `INSERT INTO user_wallets (id, user_id, currency, balance) VALUES ($1, $2, $3, $4)`
	for i, exchange := range currExch {
		if err := s.db.Exec(query, iduuid[i], idUser, exchange.Currency, 0).Error; err != nil {
			return fmt.Errorf("Ошибка создания кошелька %s для пользователя: %v", exchange.Currency, err)
		}
	}

	// Логируем успешную вставку
	log.Printf("Кошельки успешно добавлены для пользователя с ID: %s\n", idUser)
	return nil
}

func GetBalanceAfterOperation(db *gorm.DB, ctx context.Context, userID uuid.UUID) (map[string]float32, error) {
	var wallets []models.UserWallet
	if err := db.Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("Ошибка получения баланса: %v", err)
	}

	balances := make(map[string]float32)
	for _, wallet := range wallets {
		balances[wallet.Currency] = wallet.Balance
	}

	return balances, nil
}

func (s *Storage) GetBalance(ctx context.Context, token string) (map[string]float32, error) {

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("Ошибка в валидации токена: %v", err)
	}

	userID := claims.UserID

	balances, err := GetBalanceAfterOperation(s.db, ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	return balances, nil
}

func (s *Storage) Withdraw(ctx context.Context, token string, amount float32, currency string) (string, map[string]float32, error) {

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка в валидации токена: %v", err)
	}

	currExch, err := GetExchangeRates(s.db)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка получения списка валют: %v", err)
	}

	validCurrencies := map[string]bool{}
	for _, exchange := range currExch {
		validCurrencies[exchange.Currency] = true
	}

	if !validCurrencies[currency] {
		return "", nil, fmt.Errorf("Неверная валюта: %s", currency)
	}

	if amount <= 0 {
		return "", nil, fmt.Errorf("Сумма должна быть больше нуля, запрашиваемая сумма %.2f", amount)
	}

	userID := claims.UserID
	var wallet UserWallet
	err = s.db.FirstOrCreate(&wallet, UserWallet{UserID: userID, Currency: currency}).Error
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка получения кошелька: %w", err)
	}

	if wallet.Balance < amount {
		return "", nil, fmt.Errorf("недостаточно средств на счете: текущий баланс %.2f %s, запрашиваемая сумма %.2f %s", wallet.Balance, currency, amount, currency)
	}

	wallet.Balance -= amount
	if err := s.db.Save(&wallet).Error; err != nil {
		return "", nil, fmt.Errorf("Ошибка обновления баланса: %w", err)
	}
	newBalance, err := GetBalanceAfterOperation(s.db, ctx, userID)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка получения баланса: %w", err)
	}

	return "Withdrawal successful", newBalance, nil
}

func (s *Storage) Deposit(ctx context.Context, token string, amount float32, currency string) (string, map[string]float32, error) {

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка в валидации токена: %v", err)
	}

	userID := claims.UserID
	if amount <= 0 {
		return "", nil, fmt.Errorf("Сумма должна быть больше нуля")
	}

	currExch, err := GetExchangeRates(s.db)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка получения списка валют: %v", err)
	}

	validCurrencies := map[string]bool{}
	for _, exchange := range currExch {
		validCurrencies[exchange.Currency] = true
	}

	if !validCurrencies[currency] {
		return "", nil, fmt.Errorf("Неверная валюта: %s", currency)
	}

	var wallet UserWallet
	err = s.db.FirstOrCreate(&wallet, UserWallet{UserID: userID, Currency: currency}).Error
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка получения кошелька: %w", err)
	}

	wallet.Balance += amount
	if err := s.db.Save(&wallet).Error; err != nil {
		return "", nil, fmt.Errorf("Ошибка обновления баланса: %w", err)
	}
	newBalance, err := GetBalanceAfterOperation(s.db, ctx, userID)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка получения баланса: %w", err)
	}

	return "Account topped up successfully", newBalance, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("Ошибка получения пользователя: %v", err)
	}
	return user, nil
}

func (s *Storage) GetExchangeRates(ctx context.Context, token string) (string, map[string]float32, error) {

	_, err := jwt.ValidateToken(token)
	if err != nil {
		return "", nil, fmt.Errorf("Ошибка в валидации токена: %v", err)
	}

	// Извлечение всех курсов валют из базы данных
	var exchangeRates []ExchangeRate
	if err := s.db.Find(&exchangeRates).Error; err != nil {
		return "", nil, fmt.Errorf("не удалось получить курсы валют: %w", err)
	}

	// Формирование карты курсов валют
	rates := make(map[string]float32)
	for _, rate := range exchangeRates {
		rates[rate.Currency] = rate.RateToUSD
	}

	return "Курсы валют успешно получены", rates, nil
}

func (s *Storage) ExchangeCurrency(ctx context.Context, token string,
	from_currency string, to_currency string, amount float32) (string, float32, map[string]float32, error) {

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		return "", 0, nil, fmt.Errorf("Ошибка в валидации токена: %v", err)
	}

	if amount <= 0 {
		return "", 0, nil, fmt.Errorf("сумма обмена должна быть больше нуля")
	}

	// Обновляем курсы валют при обмене
	err = UpdateExchangeRates(s.db)
	if err != nil {
		log.Printf("Ошибка обновления курсов валют: %v", err)
	} else {
		log.Println("Курсы валют успешно обновлены.")
	}

	currExch, err := GetExchangeRates(s.db)
	if err != nil {
		return "", 0, nil, fmt.Errorf("Ошибка получения списка валют: %v", err)
	}

	validCurrencies := map[string]bool{}
	for _, exchange := range currExch {
		validCurrencies[exchange.Currency] = true
	}

	if !validCurrencies[from_currency] || !validCurrencies[to_currency] {
		return "", 0, nil, fmt.Errorf("недопустимые валюты: %s или %s", from_currency, to_currency)
	}

	userID := claims.UserID

	// Получение кошельков пользователя
	var fromWallet, toWallet UserWallet
	if err := s.db.FirstOrCreate(&fromWallet, UserWallet{UserID: userID, Currency: from_currency}).Error; err != nil {
		return "", 0, nil, fmt.Errorf("не удалось найти или создать кошелек: %w", err)
	}
	if err := s.db.FirstOrCreate(&toWallet, UserWallet{UserID: userID, Currency: to_currency}).Error; err != nil {
		return "", 0, nil, fmt.Errorf("не удалось найти или создать кошелек: %w", err)
	}

	// Проверка достаточности средств для обмена
	if fromWallet.Balance < amount {
		return "", 0, nil, fmt.Errorf("недостаточно средств на счете %s: текущий баланс %.2f, запрашиваемая сумма %.2f", from_currency, fromWallet.Balance, amount)
	}

	// Получение курсов валют
	var fromRate, toRate ExchangeRate
	if err := s.db.First(&fromRate, "currency = ?", from_currency).Error; err != nil {
		return "", 0, nil, fmt.Errorf("не удалось получить курс валюты %s: %w", from_currency, err)
	}
	if err := s.db.First(&toRate, "currency = ?", to_currency).Error; err != nil {
		return "", 0, nil, fmt.Errorf("не удалось получить курс валюты %s: %w", to_currency, err)
	}

	amountInUSD := amount / fromRate.RateToUSD
	exchangedAmount := amountInUSD * toRate.RateToUSD

	fromWallet.Balance -= amount
	toWallet.Balance += exchangedAmount

	if err := s.db.Save(&fromWallet).Error; err != nil {
		return "", 0, nil, fmt.Errorf("не удалось обновить баланс %s: %w", from_currency, err)
	}
	if err := s.db.Save(&toWallet).Error; err != nil {
		return "", 0, nil, fmt.Errorf("не удалось обновить баланс %s: %w", to_currency, err)
	}

	// Получение нового баланса
	newBalance, err := GetBalanceAfterOperation(s.db, ctx, userID)
	if err != nil {
		return "", 0, nil, fmt.Errorf("не удалось получить новый баланс: %w", err)
	}

	return "Обмен успешно завершен", exchangedAmount, newBalance, nil
}
