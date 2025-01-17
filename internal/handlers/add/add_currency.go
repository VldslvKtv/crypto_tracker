package add

import (
	"context"
	"crypto_tracker/internal/models"
	"crypto_tracker/internal/tracker"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type AddNewCoin interface {
	AddCoin(ctx context.Context, coin models.Coin) error
}

// @Summary Добавить криптовалюту для отслеживания
// @Description Добавляет криптовалюту в список отслеживаемых и начинает сбор данных о её цене.
// @ID add-coin
// @Accept json
// @Produce json
// @Param request body models.CoinRequest true "Данные для добавления криптовалюты"
// @Success 200 {object} map[string]string "message: Currency added to watchlist"
// @Failure 400 {object} map[string]string "error: Invalid request body"
// @Failure 400 {object} map[string]string "error: Invalid coin"
// @Failure 400 {object} map[string]string "error: Coin is already being tracked"
// @Router /currency/add [post]
func New(log *slog.Logger, apiURL string, apiKey string, addNewCoin AddNewCoin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсим JSON
		var req models.CoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid request body"})
			return
		}

		// Проверяем, существует ли валюта через внешний API
		if !isValidCoin(apiURL, apiKey, req.Coin) {
			log.Warn("Invalid coin", "coin", req.Coin)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid coin"})
			return
		}

		// Проверяем, не отслеживается ли уже эта криптовалюта
		tracker.TrackedMutex.Lock()
		if _, exists := tracker.TrackedCoins[req.Coin]; exists {
			tracker.TrackedMutex.Unlock()
			log.Warn("Coin is already being tracked", "coin", req.Coin)
			render.JSON(w, r, map[string]string{"error": "Coin is already being tracked"})
			return
		}

		// Добавляем криптовалюту в мапу отслеживаемых
		tracker.TrackedCoins[req.Coin] = true
		tracker.TrackedMutex.Unlock()

		// Запускаем горутину для сбора данных
		go func() {
			startPriceCollector(r.Context(), log, addNewCoin, apiURL, apiKey, req.Coin)
		}()

		// Сообщаем, что валюта добавлена на наблюдение
		render.JSON(w, r, map[string]string{"message": "Currency added to watchlist"})
	}
}

func isValidCoin(apiURL, apiKey, coin string) bool {
	url := fmt.Sprintf("%s/api/1/metadata?asset=%s&api_key=%s", apiURL, coin, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Если статус ответа не 200, возвращаем false
	return resp.StatusCode == http.StatusOK

}

func startPriceCollector(ctx context.Context, log *slog.Logger, addNewCoin AddNewCoin, apiURL, apiKey, coin string) {
	// Создаем канал для остановки горутины
	stopChan := make(chan struct{})

	// Сохраняем канал в глобальной мапе
	tracker.StopMutex.Lock()
	tracker.StopChannels[coin] = stopChan
	tracker.StopMutex.Unlock()

	// Удаляем канал из мапы при завершении горутины
	defer func() {
		tracker.TrackedMutex.Lock()
		delete(tracker.TrackedCoins, coin)
		tracker.TrackedMutex.Unlock()
	}()

	ticker := time.NewTicker(10 * time.Second) // Интервал сбора данных
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			// Останавливаем горутину, если получен сигнал
			log.Info("Stopped price collection for coin", "coin", coin)
			return
		case <-ticker.C:
			info, err := fetchPriceFromAPI(apiURL, apiKey, coin)
			if err != nil {
				log.Warn("Failed to fetch price", "coin", coin, "error", err)
				continue
			}

			// Сохраняем цену в базу данных
			ctx = context.WithoutCancel(ctx)
			if err := addNewCoin.AddCoin(ctx, info); err != nil {
				log.Error("Failed to save price", "coin", coin, "error", err)
			}
		}
	}
}

func fetchPriceFromAPI(apiURL, apiKey, coin string) (models.Coin, error) {
	currentTimeMillis := time.Now().Add(-24*time.Hour).Unix() * 1000

	url := fmt.Sprintf("%s/api/1/market/history?asset=%s&from=%d&api_key=%s", apiURL, coin, currentTimeMillis,
		apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return models.Coin{}, fmt.Errorf("error: %s", err)
	}
	defer resp.Body.Close()

	var responseAPI struct {
		Data struct {
			Name         string       `json:"name"`
			PriceHistory [][2]float64 `json:"price_history"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseAPI); err != nil {
		return models.Coin{}, err
	}

	log.Printf("responseAPI getting %v", responseAPI.Data.PriceHistory[len(responseAPI.Data.PriceHistory)-1])

	if price := responseAPI.Data.PriceHistory[0][1]; price != 0 {
		coinInfo := models.Coin{
			Name:      responseAPI.Data.Name,
			Price:     responseAPI.Data.PriceHistory[len(responseAPI.Data.PriceHistory)-1][1],
			Timestamp: int64(responseAPI.Data.PriceHistory[len(responseAPI.Data.PriceHistory)-1][0]),
		}
		return coinInfo, nil
	}

	return models.Coin{}, fmt.Errorf("price not found for %s", coin)
}
