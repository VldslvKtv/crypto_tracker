package pg

import (
	"context"
	"crypto_tracker/config"
	"crypto_tracker/internal/models"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.pg.New"

	databaseUrl := cfg.StoragePath
	dbPool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("%s :%w", op, err)
	}
	if err := runMigrations(cfg); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	return &Storage{DB: dbPool}, nil

}

func runMigrations(cfg *config.Config) error {
	const op = "storage.pg.runMigrations"
	m, err := migrate.New(cfg.MigrationsPath, cfg.StoragePath)
	if err != nil {
		return fmt.Errorf("%s :%w", op, err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	return nil

}
func (s *Storage) Close() {
	defer s.DB.Close()
}

func (s *Storage) AddCoin(ctx context.Context, coin models.Coin) error {
	const op = "storage.pg.AddCoin"
	_, err := s.DB.Exec(ctx, `
        INSERT INTO coins (name, price, fixation_time)
        VALUES ($1, $2, $3)
    `, coin.Name, coin.Price, coin.Timestamp)
	if err != nil {
		return fmt.Errorf("%s; failed to insert coin: %w", op, err)
	}
	return nil
}

func (s *Storage) GetPrice(ctx context.Context, coin string, timestamp int64) (models.Coin, error) {
	const op = "storage.pg.GetPrice"
	var coinInfo models.Coin
	err := s.DB.QueryRow(ctx, `
        SELECT name, price, fixation_time
        FROM coins
        WHERE name = $1 AND fixation_time <= $2
        ORDER BY ABS(fixation_time - $2)
        LIMIT 1
    `, coin, timestamp).Scan(&coinInfo.Name, &coinInfo.Price, &coinInfo.Timestamp)
	if err != nil {
		return models.Coin{}, fmt.Errorf("%s; failed to get coin: %w", op, err)
	}
	return coinInfo, nil
}
