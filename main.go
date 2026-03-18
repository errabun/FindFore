package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ericrabun/findfore-go/internal/config"
	"github.com/ericrabun/findfore-go/internal/database"
	"github.com/ericrabun/findfore-go/internal/handler"
	"github.com/ericrabun/findfore-go/internal/router"
	"github.com/ericrabun/findfore-go/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	queries := store.New(db)
	h := handler.New(queries, db, cfg.JWTSecret)
	r := router.New(h, cfg.JWTSecret)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
