package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gh0xFF/rates/pkg/model"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func (s *RatesService) startCronJob(ctx context.Context) {
	logrus.Info("job started")
	location, err := time.LoadLocation("UTC")
	if err != nil {
		logrus.Errorf("Error loading location: %q", err)
		panic(err)
	}

	rates, err := s.getTodaysRates(ctx)
	if err != nil {
		logrus.Errorf("Error while getting rates: %q", err)

		time.Sleep(5 * time.Second)

		rates, err = s.getTodaysRates(ctx)
		if err != nil {
			logrus.Errorf("Error while getting rates: %q", err)
			panic(err)
		}
	}

	data, err := s.storage.GetRatesByDate(ctx, time.Now())
	if err != nil {
		logrus.Errorf("Error while chicking todays rates in database")
	}

	if len(data) == 0 {
		if err = s.storage.InsertDayRates(ctx, rates); err != nil {
			logrus.Errorf("Error while storing rates to db: %q", err)
			panic(err)
		}
	}

	s.lock.Lock()
	s.rates = rates
	s.lock.Unlock()

	logrus.Info("Starting cron job")

	// Create a new cron scheduler with the specified timezone
	c := cron.New(cron.WithLocation(location))

	// каждый день в 12:00
	_, err = c.AddFunc("00 12 * * *", func() {
		rates, err := s.getTodaysRates(ctx)
		if err != nil {
			logrus.Errorf("cron: error while getting rates: %q", err)
			logrus.Warn("waiting 5 sec and retrying to get rates")

			time.Sleep(5 * time.Second)

			rates, err = s.getTodaysRates(ctx)
			if err != nil {
				logrus.Errorf("cron: error while getting rates: %q", err)
				panic("rates dosn't loaded after retry, check if host available, error: " + err.Error())
			}
		}

		s.lock.RLock()
		s.rates = rates
		s.lock.RUnlock()
	})

	if err != nil {
		logrus.Errorf("Error scheduling task: %q", err)
		panic(err)
	}

	c.Start() // Start the cron scheduler
	select {} // Keep the program running
}

func (s *RatesService) getTodaysRates(ctx context.Context) ([]model.RateModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.ratesUrl, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if rsp.Status != "200 OK" {
		return nil, errors.New("expected status 200, got " + rsp.Status)
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	var rates []model.RateModel
	if err = json.Unmarshal(data, &rates); err != nil {
		return nil, err
	}

	return rates, nil
}
