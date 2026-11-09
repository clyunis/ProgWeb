// Package handlers contiene los handlers HTTP de la aplicación.
// Cada handler obtiene los datos con sqlc y renderiza una vista templ (HTML),
// siguiendo el modelo de las filminas de SSR: Handler -> SQLC Queries -> Templ.
package handlers

import (
	db "tp5/db/sqlc"
)

// Handler agrupa los handlers de la app y les da acceso a la base de datos.
// En lugar de usar una variable global, las queries se reciben al crearlo.
type Handler struct {
	queries *db.Queries
}

// New crea el Handler a partir de las queries generadas por sqlc.
func New(q *db.Queries) *Handler {
	return &Handler{queries: q}
}
