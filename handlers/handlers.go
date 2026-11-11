package handlers

import (
	db "tp5/db/sqlc"
)

type Handler struct {
	queries *db.Queries
}

func New(q *db.Queries) *Handler {
	return &Handler{queries: q}
}
