package remove

import (
	"crypto_tracker/internal/models"
	"crypto_tracker/internal/tracker"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

// @Summary Удалить криптовалюту из отслеживаемых
// @Description Удаляет криптовалюту из списка отслеживаемых и останавливает сбор данных о её цене.
// @ID remove-coin
// @Accept json
// @Produce json
// @Param request body models.CoinRequest true "Данные для удаления криптовалюты"
// @Success 200 {object} map[string]string "message: Currency removed from watchlist"
// @Failure 400 {object} map[string]string "error: Invalid request body"
// @Failure 400 {object} map[string]string "error: Coin field is required"
// @Failure 404 {object} map[string]string "error: Coin is not tracked"
// @Router /currency/remove [post]
func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсим JSON
		var req models.CoinRequest
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

		stopPriceCollector(log, req.Coin)

		render.JSON(w, r, map[string]string{"message": "Currency removed from watchlist"})
	}
}

func stopPriceCollector(log *slog.Logger, coin string) {
	// Блокируем доступ к мапе
	tracker.StopMutex.Lock()
	defer tracker.StopMutex.Unlock()

	// Останавливаем горутину, если канал существует
	if stopChan, ok := tracker.StopChannels[coin]; ok {
		close(stopChan)                    // Отправляем сигнал остановки
		delete(tracker.StopChannels, coin) // Удаляем канал из мапы
		log.Info("Stopped price collection for coin", "coin", coin)
		log.Info("Map: ", "map", coin)
	}
}
