package postgresql

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type ExchangeRateAPI struct {
	Rates map[string]float64 `json:"rates"`
}

func UpdateExchangeRates(db *gorm.DB) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Отключение проверки сертификата
	}
	client := &http.Client{Transport: tr}

	url := "https://api.exchangerate-api.com/v4/latest/USD?symbols=RUB,EUR"
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Ошибка запроса к API: %v", err)
	}
	defer resp.Body.Close()

	var data ExchangeRateAPI
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("Ошибка декодирования ответа: %v", err)
	}

	neededCurrencies := []string{"USD", "RUB", "EUR"}

	query := `
		INSERT INTO exchange_rates (currency, rate_to_usd) 
		VALUES ($1, $2)
		ON CONFLICT (currency) 
		DO UPDATE SET rate_to_usd = EXCLUDED.rate_to_usd`

	for _, currency := range neededCurrencies {
		if rate, found := data.Rates[currency]; found {
			fmt.Printf("Курс для %s: %.4f\n", currency, rate)
			if err := db.Exec(query, currency, float32(rate)).Error; err != nil {
				return fmt.Errorf("Ошибка добавления или обновления курса для %s: %v", currency, err)
			}
		} else {
			fmt.Printf("Курс для %s не найден\n", currency)
		}
	}

	return nil
}
