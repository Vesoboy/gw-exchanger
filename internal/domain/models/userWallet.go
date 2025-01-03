package models

import "github.com/google/uuid"

type UserWallet struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey"`
	UserID   uuid.UUID `json:"user_id" gorm:"index"`                                                           // Внешний ключ на пользователя
	Currency string    `json:"currency"`                                                                       // Валюта (USD, EUR, RUB)
	Balance  float32   `json:"balance" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Баланс в конкретной валюте
}
