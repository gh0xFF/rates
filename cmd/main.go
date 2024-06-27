package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gh0xFF/rates/pkg/data"
	"github.com/gh0xFF/rates/pkg/handler"
	"github.com/gh0xFF/rates/pkg/service"

	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	logrus.SetLevel(logrus.WarnLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var connectionString = os.Getenv("CONNECTION_STRING")
	if connectionString == "" {
		logrus.Fatal("CONNECTION_STRING is not set")
	}

	var ratesUrl = os.Getenv("RATES_URL")
	if ratesUrl == "" {
		logrus.Fatal("RATES_URL is not set")
	}

	var port = os.Getenv("PORT")
	if port == "" {
		logrus.Fatal("PORT is not set")
	}

	var env = os.Getenv("ENV")
	if env == "" {
		logrus.Fatal("ENV is not set")
	}

	// var env = "prod"
	// var port = "8080"
	// var connectionString = "root:password@tcp(localhost:3306)/rates"
	// var ratesUrl = `https://api.nbrb.by/exrates/rates?periodicity=0`

	data, err := data.NewData(ctx, connectionString, env)
	if err != nil {
		logrus.Fatalf("failed to initialize data layer: %s", err.Error())
	}

	handlers := handler.NewHandler(
		service.NewService(data, ratesUrl),
	)

	srv := new(server)
	go func() {
		if err := srv.Run(port, handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Printf("github.com/gh0xFF/rates Shutting Down")

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := data.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}
}

type server struct {
	httpSrv *http.Server
}

func (s *server) Run(port string, h http.Handler) error {
	s.httpSrv = &http.Server{
		Addr:           ":" + port,
		Handler:        h,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
	}

	return s.httpSrv.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
