package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type ClientsRepoInterface interface {
	GetOrCreate(ctx context.Context, key string, capacity int32, ratePerSec int) (int32, int, error)
}

type ClientsRepo struct {
	db *sql.DB
}

func NewClientsRepo(db *sql.DB) *ClientsRepo {
	return &ClientsRepo{db: db}
}

func (cr *ClientsRepo) GetOrCreate(ctx context.Context, key string, capacity int32, ratePerSec int) (int32, int, error) {
	var cap int32
	var rate int
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return cap, rate, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, "SELECT capacity, rate_per_sec FROM clients WHERE id = $1", key).Scan(&cap, &rate)
	if err != nil && err != sql.ErrNoRows {
		return cap, rate, err
	} else if err == nil {
		return cap, rate, nil
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO clients (id, capacity, rate_per_sec) VALUES ($1, $2, $3)", key, capacity, ratePerSec)
	if err != nil {
		return cap, rate, err
	}

	if err := tx.Commit(); err != nil {
		return cap, rate, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return capacity, ratePerSec, nil
}
