package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	db "tp3/db/sqlc"
)

func publicacionesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/publicaciones")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			getPublicaciones(w, r)
		case http.MethodPost:
			createPublicacion(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de publicación inválido"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getPublicacion(w, r, int32(id))
	case http.MethodPut:
		updatePublicacion(w, r, int32(id))
	case http.MethodDelete:
		deletePublicacion(w, r, int32(id))
	default:
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
	}
}

// GET /publicaciones
func getPublicaciones(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pubs, err := queries.ListPublicaciones(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al consultar la base de datos"})
		return
	}

	if pubs == nil {
		pubs = []db.Publicacion{} // Retorna array vacío en lugar de null si no hay registros
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pubs)
}

// POST /publicaciones
func createPublicacion(w http.ResponseWriter, r *http.Request) {
	var params db.CreatePublicacionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	// Validación simple
	if strings.TrimSpace(params.Titulo) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El título es obligatorio"})
		return
	}

	ctx := r.Context()
	idGenerado, err := queries.CreatePublicacion(ctx, params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al guardar en la base de datos"})
		return
	}

	// Recuperar y devolver la entidad recién creada
	nuevaPub, _ := queries.GetPublicacion(ctx, idGenerado)
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(nuevaPub)
}

// GET /publicaciones/{id}
func getPublicacion(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()
	pub, err := queries.GetPublicacion(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Publicación no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pub)
}

// PUT /publicaciones/{id}
func updatePublicacion(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	var params db.UpdatePublicacionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	if strings.TrimSpace(params.Titulo) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El título es obligatorio"})
		return
	}

	params.IDAlquiler = id

	if _, err := queries.GetPublicacion(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Publicación no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.UpdatePublicacion(ctx, params); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al actualizar la publicación"})
		return
	}

	pubActualizada, _ := queries.GetPublicacion(ctx, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pubActualizada)
}

// DELETE /publicaciones/{id}
func deletePublicacion(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	if _, err := queries.GetPublicacion(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Publicación no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.DeletePublicacion(ctx, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al eliminar"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
