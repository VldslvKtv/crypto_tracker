package remove

import (
	"crypto_tracker/internal/tracker"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсим JSON
		var req struct {
			Coin string `json:"coin"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid request body"})
			return
		}

		// Проверяем, что поле "coin" не пустое
		if strings.TrimSpace(req.Coin) == "" {
			log.Error("Empty coin field")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Coin field is required"})
			return
		}

		//Проверяем, отслеживается ли эта криптовалюта
		tracker.TrackedMutex.Lock()
		if _, exists := tracker.TrackedCoins[req.Coin]; !exists {
			tracker.TrackedMutex.Unlock()
			log.Warn("Coin is not tracked", "coin", req.Coin)
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]string{"error": "Coin is not tracked"})
			return
		}

		tracker.TrackedMutex.Unlock()

		stopPriceCollector(req.Coin)

		render.JSON(w, r, map[string]string{"message": "Currency removed from watchlist"})
	}
}

func stopPriceCollector(coin string) {
	// Блокируем доступ к мапе
	tracker.StopMutex.Lock()
	defer tracker.StopMutex.Unlock()

	// Останавливаем горутину, если канал существует
	if stopChan, ok := tracker.StopChannels[coin]; ok {
		close(stopChan)                    // Отправляем сигнал остановки
		delete(tracker.StopChannels, coin) // Удаляем канал из мапы
		slog.Info("Stopped price collection for coin", "coin", coin)
		slog.Info("Map: ", "map", coin)
	}
}
