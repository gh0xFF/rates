package data

import (
	"context"
	"errors"
	"time"

	"github.com/gh0xFF/rates/pkg/model"

	_ "github.com/lib/pq"
)

type Storage interface {
	InsertDayRates(ctx context.Context, p []model.RateModel) error
	GetRatesByDate(ctx context.Context, date time.Time) ([]model.RateModel, error)
	GetRatesDump(ctx context.Context) ([]model.RateModel, error)

	Ping(ctx context.Context) error
	Close() error
}

func NewData(ctx context.Context, connectionString, env string) (Storage, error) {
	if env == "local" { // to run without real db
		return NewMockData(ctx)
	}

	if env == "prod" {
		return NewMysqlDB(ctx, connectionString)
	}

	return nil, errors.New("unknown env: " + env)
}
