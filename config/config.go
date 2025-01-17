package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	StoragePath    string
	MigrationsPath string
	HTTPServer
	APIUrls
}

type HTTPServer struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

type APIUrls struct {
	ExtAPIUrl string
	APIKey    string
}

func MustLoad() Config {

	// Загружаем переменные окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	// Читаем переменные окружения и заполняем структуру
	config := Config{
		Env:            checkAndReturnData("ENV"),
		StoragePath:    checkAndReturnData("DATABASE_URL"),
		MigrationsPath: checkAndReturnData("MIGRATIONS"),
		HTTPServer: HTTPServer{
			Address:     checkAndReturnData("HTTP_SERVER_ADDRESS"),
			Timeout:     parseDuration(os.Getenv("HTTP_SERVER_TIMEOUT")),
			IdleTimeout: parseDuration(os.Getenv("HTTP_SERVER_IDLE_TIMEOUT")),
		},
		APIUrls: APIUrls{
			ExtAPIUrl: checkAndReturnData("API_URL"),
			APIKey:    checkAndReturnData("API_KEY"),
		},
	}

	log.Printf("Config: %+v\n", config)
	return config
}

// Проверка, что поля не пустые
func checkAndReturnData(s string) string {
	data := os.Getenv(s)
	if data == "" {
		log.Fatalf("Поле %s пустое", s)
	}
	return data
}

// Преобразование строки во временной интервал
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("Error parsing duration: %v", err)
	}
	return d
}
