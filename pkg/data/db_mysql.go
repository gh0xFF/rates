package data

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gh0xFF/rates/pkg/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Data struct {
	db *sqlx.DB
}

func NewMysqlDB(ctx context.Context, connectionString string) (Storage, error) {
	db, err := sqlx.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(10 * time.Second)
	db.SetConnMaxLifetime(10 * time.Second)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

	data := &Data{db}

	// схема миграции кривая, но рабочая
	// тратить время на настройку тулов типа liquidbase нет желания
	if err = data.runMigration(ctx); err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Data) runMigration(ctx context.Context) error {
	schema := `CREATE TABLE IF NOT EXISTS rates (
	date            		DATE 		 NOT NULL,
	currency_id           	INT(128)	 NOT NULL,
	currency_scale          INT(128)   	 NOT NULL,
	currency_official_rate  FLOAT(16) 	 NOT NULL
)`

	res, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err != nil {
		return err
	} else {
		if n > 0 { // подобный индекс позволит обеспеить уникальность данных и избавит нас от дубликатов
			index := `CREATE UNIQUE INDEX rates_date_idx ON rates (date, currency_id)`
			if _, err = r.db.ExecContext(ctx, index); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Data) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Data) Close() error {
	return r.db.Close()
}

func (r *Data) InsertDayRates(ctx context.Context, data []model.RateModel) error {
	const query = `INSERT INTO rates
	(date, currency_id, currency_scale, currency_official_rate)
	VALUES(?, ?, ?, ?)`

	// не лучшая идея записывать по одному объекту, но у 32 запроса на запись в сутки
	var err error
	var missed = 0

	for _, v := range data {
		if _, err = r.db.ExecContext(ctx, query, v.Date, v.CurrencyId, v.CurrencyScale, v.CurrencyOfficialRate); err != nil {
			missed++
		}
	}

	if missed > 0 {
		return errors.New("has " + strconv.Itoa(missed) + "errors while insert data, error: " + err.Error())
	}

	return nil
}

func (r *Data) GetRatesByDate(ctx context.Context, date time.Time) ([]model.RateModel, error) {
	var query = `SELECT date, currency_id, currency_scale, currency_official_rate
	FROM rates.rates WHERE date = ?`

	var p []model.RateModel
	rows, err := r.db.QueryContext(ctx, query, date.In(time.Local).Format("2006-01-02"))
	if err != nil {
		return p, err
	}

	defer rows.Close()

	for rows.Next() {
		var m model.RateModel
		if err = rows.Scan(&m.Date, &m.CurrencyId, &m.CurrencyScale, &m.CurrencyOfficialRate); err != nil {
			return p, err
		}

		p = append(p, m)
	}

	return p, nil
}

func (r *Data) GetRatesDump(ctx context.Context) ([]model.RateModel, error) {
	const query = `SELECT date, currency_id, currency_scale, currency_official_rate
	FROM rates ORDER BY date DESC`

	var p []model.RateModel
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return p, err
	}

	defer rows.Close()

	for rows.Next() {
		var t model.RateModel
		if err = rows.Scan(&t.Date, &t.CurrencyId, &t.CurrencyScale, &t.CurrencyOfficialRate); err != nil {
			return p, err
		}

		p = append(p, t)
	}

	return p, nil
}
