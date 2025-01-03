package postgresql

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID    `json:"id" gorm:"primaryKey"`
	Username string       `json:"username" gorm:"unique"`
	Email    string       `json:"email" gorm:"unique"`
	PassHash []byte       `json:"password"`
	Balances []UserWallet `json:"balances" gorm:"foreignKey:UserID"`
}

type UserWallet struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey"`
	UserID   uuid.UUID `json:"user_id" gorm:"index"`                                                           // Внешний ключ на пользователя
	Currency string    `json:"currency"`                                                                       // Валюта (USD, EUR, RUB)
	Balance  float32   `json:"balance" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Баланс в конкретной валюте
}

type ExchangeRate struct {
	Currency  string  `json:"currency" gorm:"primaryKey"` // Валюта (например, USD)
	RateToUSD float32 `json:"rate_to_usd"`                // Курс относительно базовой валюты USD
}
