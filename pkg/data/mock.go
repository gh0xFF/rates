package data

import (
	"context"
	"time"

	"github.com/gh0xFF/rates/pkg/model"
)

// этот файл мне нужен чтобы убедиться в работоспособности сервиса не поднимая mysql

type Key struct {
	CurrencyId uint32
	Date       string
}

type MockData struct {
	data map[Key]model.RateModel
}

func NewMockData(ctx context.Context) (Storage, error) {
	var data = make(map[Key]model.RateModel)
	return &MockData{data: data}, nil
}

func (m *MockData) Ping(ctx context.Context) error {
	return nil
}

func (m *MockData) Close() error {
	return nil
}

func (m *MockData) InsertDayRates(ctx context.Context, d []model.RateModel) error {
	for _, v := range d {
		v.CurrencyAbbrevation = ""
		v.CurrencyName = ""

		m.data[Key{CurrencyId: v.CurrencyId, Date: v.Date}] = v
	}
	return nil
}

func (m *MockData) GetRatesByDate(ctx context.Context, date time.Time) ([]model.RateModel, error) {
	var data = make([]model.RateModel, 0, 31)
	for _, v := range m.data {
		if v.Date[0:10] == date.Format("2006-01-02") { // криво, но для теста ок
			data = append(data, v)
		}
	}

	return data, nil
}

func (m *MockData) GetRatesDump(ctx context.Context) ([]model.RateModel, error) {
	var data = make([]model.RateModel, 0, len(m.data))
	for _, v := range m.data {
		data = append(data, v)
	}

	return data, nil
}
