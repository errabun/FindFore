package handler

import (
	"database/sql"

	"github.com/ericrabun/findfore-go/internal/store"
)

type Handler struct {
	queries   *store.Queries
	db        *sql.DB
	jwtSecret string
}

func New(queries *store.Queries, db *sql.DB, jwtSecret string) *Handler {
	return &Handler{
		queries:   queries,
		db:        db,
		jwtSecret: jwtSecret,
	}
}
