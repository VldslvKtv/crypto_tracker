package get

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"crypto_tracker/internal/models"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type PriceStorage interface {
	GetPrice(ctx context.Context, coin string, timestamp int64) (models.Coin, error)
}

func New(log *slog.Logger, storage PriceStorage) http.HandlerFunc {
	validate := validator.New() // Создаем экземпляр валидатора

	return func(w http.ResponseWriter, r *http.Request) {
		//Парсим JSON
		var req struct {
			Coin      string `json:"coin" validate:"required"`
			Timestamp string `json:"timestamp" validate:"required"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid request body"})
			return
		}

		// Валидируем запрос
		if err := validate.Struct(req); err != nil {
			log.Error("Validation failed", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Validation failed: coin and timestamp are required"})
			return
		}

		// Получаем цену из базы данных
		timestamp, err := strconv.ParseInt(req.Timestamp, 10, 64)
		if err != nil {
			log.Error("Failed to get timestamp", "coin", req.Coin, "timestamp", req.Timestamp, "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Failed to get price"})
			return
		}
		coinInfo, err := storage.GetPrice(r.Context(), req.Coin, timestamp)
		if err != nil {
			log.Error("Failed to get price", "coin", req.Coin, "timestamp", req.Timestamp, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Failed to get price"})
			return
		}

		response := models.Coin{
			Name:      coinInfo.Name,
			Price:     coinInfo.Price,
			Timestamp: coinInfo.Timestamp,
		}
		render.JSON(w, r, response)
	}
}
