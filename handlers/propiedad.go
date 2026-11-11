package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tp5/views"
)

func (h *Handler) Propiedad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	usuario := h.usuarioActual(r)

	idStr := strings.TrimPrefix(r.URL.Path, "/propiedad/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render(w, r, http.StatusNotFound, views.NoEncontradaPage(usuario))
		return
	}

	ctx := r.Context()
	pub, err := h.queries.GetPublicacionDetalle(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		render(w, r, http.StatusNotFound, views.NoEncontradaPage(usuario))
		return
	}
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}

	fotos, err := h.queries.ListImagenesByPublicacion(ctx, int32(id))
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	contacto, err := h.queries.GetContactoPublicacion(ctx, int32(id))
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}

	actual, err := strconv.Atoi(r.URL.Query().Get("foto"))
	if err != nil || actual < 0 || actual >= len(fotos) {
		actual = 0
	}

	render(w, r, http.StatusOK, views.PublicacionPage(pub, fotos, actual, contacto, usuario))
}
