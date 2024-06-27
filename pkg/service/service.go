package service

import (
	"context"
	"sync"
	"time"

	"github.com/gh0xFF/rates/pkg/data"
	"github.com/gh0xFF/rates/pkg/model"
)

type RatesService struct {
	ratesUrl string
	rates    []model.RateModel
	lock     sync.RWMutex // мьютекс должен срабатывать раз в сутки
	storage  data.Storage
}

func NewRatesService(repo data.Storage, ratesUrl string) *RatesService {
	s := &RatesService{
		ratesUrl: ratesUrl,
		storage:  repo,
		lock:     sync.RWMutex{},
		rates:    make([]model.RateModel, 0, 31),
	}

	go s.startCronJob(context.Background())
	return s
}

func (s *RatesService) GetRatesByDate(ctx context.Context, date time.Time) ([]model.RateModel, error) {
	data, err := s.storage.GetRatesByDate(ctx, date)
	if err != nil {
		return data, err
	}

	for i, v := range data {
		data[i].CurrencyAbbrevation = currencyIdToCurrencyAbbreviation[v.CurrencyId]
		data[i].CurrencyName = currencyIdToCurrencyName[v.CurrencyId]
	}

	return data, nil
}

func (s *RatesService) GetTodaysRate(_ctx context.Context) []model.RateModel {
	// на самом деле непонятна критичность получения самых актуальных данных
	// если критично, то нужно раскоментить строки ниже

	// var data = make([]model.RateModel, 0, 31)
	// s.lock.Lock()
	// copy(data, s.rates)
	// s.lock.Unlock()
	// return data

	return s.rates
}

func (s *RatesService) GetRatesDump(ctx context.Context) ([]model.RateModel, error) {
	data, err := s.storage.GetRatesDump(ctx)
	if err != nil {
		return data, err
	}

	for i, v := range data {
		data[i].CurrencyAbbrevation = currencyIdToCurrencyAbbreviation[v.CurrencyId]
		data[i].CurrencyName = currencyIdToCurrencyName[v.CurrencyId]
	}

	return data, nil
}

func (s *RatesService) StoreTodaysRates(ctx context.Context, rates []model.RateModel) error {
	return s.storage.InsertDayRates(ctx, rates)
}

func (s *RatesService) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
