package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"ukapp/internal/dadata"
	"ukapp/internal/layaclient"
	"ukapp/internal/maxclient"
	"ukapp/internal/repository"
	"ukapp/internal/service"
	transporthttp "ukapp/internal/transport/http"
	"ukapp/internal/ukclient"
)

func main() {
	ctx := context.Background()
	repo, err := repository.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("open store", "err", err)
		os.Exit(1)
	}
	defer repo.Close()

	botToken := os.Getenv("MAX_BOT_TOKEN")
	svc := service.New(repo, maxclient.NewClient(botToken), dadata.NewClient(os.Getenv("DADATA_TOKEN")),
		ukclient.New(os.Getenv("UK_BASE_URL"), os.Getenv("UK_API_TOKEN")), os.Getenv("MAX_BOT_NAME"))
	if key := os.Getenv("UK_JWT_PRIVATE_KEY"); key != "" {
		if err := svc.UseTokenKey(key); err != nil {
			slog.Error("uk jwt key", "err", err)
			os.Exit(1)
		}
	} else {
		slog.Warn("UK_JWT_PRIVATE_KEY not set: ephemeral key, dispatcher sessions reset on restart")
	}
	var cls service.Classifier
	if u := os.Getenv("LAYA_URL"); u != "" {
		cls = layaclient.New(u)
	}
	svc.WithClassifier(cls, envFloat("LAYA_CATEGORY_THRESHOLD", 0.7))
	go svc.RunOutboxWorker(ctx)
	go svc.RunUkSyncWorker(ctx)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := transporthttp.NewServer(svc, botToken)
	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil { //nolint:gosec // таймауты — не для хакатона
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}

func envFloat(name string, def float64) float64 {
	if f, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		return f
	}
	return def
}
