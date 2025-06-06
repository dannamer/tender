package main

import (
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"

    "tender/internal/handler"
    "tender/internal/service"
    "tender/internal/storage"
)

func main() {
    addr := os.Getenv("SERVER_ADDRESS")
    if addr == "" {
        addr = "0.0.0.0:8080"
    }

    repo, err := storage.New("data.json")
    if err != nil {
        log.Fatalf("storage: %v", err)
    }

    tenderSvc := service.NewTenderService(repo)
    bidSvc := service.NewBidService(repo)

    tenderHandler := handler.NewTenderHandler(tenderSvc)
    bidHandler := handler.NewBidHandler(bidSvc)

    r := chi.NewRouter()
    r.Use(middleware.Logger)

    r.Get("/api/ping", handler.Ping)
    r.Mount("/api/tenders", tenderHandler.Routes())
    r.Mount("/api/bids", bidHandler.Routes())

    log.Printf("Starting server on %s", addr)
    if err := http.ListenAndServe(addr, r); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

