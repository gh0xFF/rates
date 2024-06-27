package service

import (
	"context"
	"time"

	"github.com/gh0xFF/rates/pkg/data"
	"github.com/gh0xFF/rates/pkg/model"
)

type RateService interface {
	GetRatesByDate(ctx context.Context, date time.Time) ([]model.RateModel, error)
	GetRatesDump(ctx context.Context) ([]model.RateModel, error)
	GetTodaysRate(ctx context.Context) []model.RateModel
	StoreTodaysRates(ctx context.Context, rates []model.RateModel) error

	Ping(ctx context.Context) error
}

func NewService(repo data.Storage, ratesUrl string) RateService {
	return NewRatesService(repo, ratesUrl)
}
