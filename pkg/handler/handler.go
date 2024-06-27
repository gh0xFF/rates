package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gh0xFF/rates/pkg/service"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service.RateService
}

func NewHandler(services service.RateService) *Handler {
	return &Handler{services}
}

func (h *Handler) InitRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	router.HandleFunc("/v1/todays_rates", h.TodaysRates).Methods(http.MethodGet)
	router.HandleFunc("/v1/rates_by_date", h.RatesByDate).Methods(http.MethodGet)
	router.HandleFunc("/v1/rates_dump", h.Dump)

	return router
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.RateService.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "text/json")
		w.Write([]byte(`{"status": "unhealthy"}`))

		logrus.Errorf("healthcheck error: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/json")
	w.Write([]byte(`{"status": "healthy"}`))
}

func (h *Handler) RatesByDate(w http.ResponseWriter, r *http.Request) {
	t, err := time.Parse("2006-01-02", r.URL.Query().Get("date"))
	if err != nil {
		logrus.Errorf("RatesByDate error while parsing time: %q", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	t1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local) // тоже открытый вопрос где будет "деплоится нац сервис"

	now := time.Now().In(time.Local)
	t2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if t1.In(time.Local).After(t2) {
		logrus.Errorf("RatesByDate error asked future rates")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("i don't know future rates)"))
		return
	}

	rates, err := h.RateService.GetRatesByDate(r.Context(), t1)
	if err != nil {
		logrus.Errorf("RatesByDate error while getting rates: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.Marshal(&rates)
	if err != nil {
		logrus.Errorf("RatesByDate error while marshalling rates: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(data)
}

func (h *Handler) TodaysRates(w http.ResponseWriter, r *http.Request) {
	rates := h.RateService.GetTodaysRate(r.Context())

	data, err := json.Marshal(&rates)
	if err != nil {
		logrus.Errorf("TodaysRates error while marshalling rates: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/json")
	w.Write(data)
}

func (h *Handler) Dump(w http.ResponseWriter, r *http.Request) {
	// тут в идеале нужна защита, чтобы слуайные пользователи не нагрузили сервис
	// также нужна пагинация, так можно легко перегрузить сервис если у него будет много данных
	// каджая структура занимает 60 байт, таких объектов 31 каждый день
	// 64(байта) * 31(курсов валюты) = 1984 байт в день
	// 1984 * 31(дней в месяце) = 61504 байт в месяц
	// 61504 * 12(месяцев в году) = 738048 байт в год(720.75 кб)
	// рассчёты показывают, что реализацию пагинации можно отложить на пару лет работы сервиса
	// редкие запросы, которые будут вызывать рост в несколько мегабайт ОЗУ не критичны
	// на случай если это часто вызываемый эндпоинт, то можно каждый день собирать дамп бд и хранить его в памяти как курс валют на сегодня
	rates, err := h.RateService.GetRatesDump(r.Context())
	if err != nil {
		logrus.Errorf("Dump error while getting rates: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	data, err := json.Marshal(&rates)
	if err != nil {
		logrus.Errorf("Dump error while marshalling rates: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	fileWriter, err := zipWriter.Create("dump.json")
	if err != nil {
		logrus.Errorf("Dump error while creating zip: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if _, err = fileWriter.Write(data); err != nil {
		logrus.Errorf("Dump error while writing to zip: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if err = zipWriter.Close(); err != nil {
		logrus.Errorf("Dump error while closing zip: %q", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Устанавливаем заголовки и отдаем архив
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"dump.zip\"")
	http.ServeContent(w, r, "dump.zip", time.Now(), bytes.NewReader(buf.Bytes()))
}
