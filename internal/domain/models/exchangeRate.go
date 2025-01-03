package models

type ExchangeRate struct {
	Currency  string  `json:"currency" gorm:"primaryKey"` // Валюта (например, USD)
	RateToUSD float32 `json:"rate_to_usd"`                // Курс относительно базовой валюты USD
}
