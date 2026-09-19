package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"mockuk/internal/api"
	"mockuk/internal/store"
)

func main() {
	st, err := store.Open(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	rate, _ := strconv.ParseFloat(os.Getenv("MOCK_FAILURE_RATE"), 64)
	r := api.NewRouter(st, os.Getenv("MOCK_API_TOKEN"), rate)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("mock-uk listening", "port", port, "failure_rate", rate)
	if err := http.ListenAndServe(":"+port, r); err != nil { //nolint:gosec // демо-сервис
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}
