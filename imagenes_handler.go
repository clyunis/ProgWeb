package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	db "tp3/db/sqlc"
)

func imagenesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/imagenes")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			getImagenesByPublicacion(w, r)
		case http.MethodPost:
			createImagen(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de imagen inválido"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getImagen(w, r, int32(id))
	case http.MethodPut:
		updateImagen(w, r, int32(id))
	case http.MethodDelete:
		deleteImagen(w, r, int32(id))
	default:
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
	}
}

// GET /imagenes?id_alquiler={id}
// Nota: no existe un ListImagenes general en queries.sql, solo
// ListImagenesByPublicacion. Por eso el listado exige este query param
// en lugar de devolver todas las imágenes de la base.
func getImagenesByPublicacion(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_alquiler")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El parámetro id_alquiler es obligatorio"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id_alquiler inválido"})
		return
	}

	ctx := r.Context()
	imagenes, err := queries.ListImagenesByPublicacion(ctx, int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al consultar la base de datos"})
		return
	}

	if imagenes == nil {
		imagenes = []db.Imagen{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(imagenes)
}

// POST /imagenes
func createImagen(w http.ResponseWriter, r *http.Request) {
	var params db.CreateImagenParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	// Validación simple
	if strings.TrimSpace(params.Url) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "La URL es obligatoria"})
		return
	}

	ctx := r.Context()
	idGenerado, err := queries.CreateImagen(ctx, params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al guardar en la base de datos"})
		return
	}

	nuevaImagen, _ := queries.GetImagen(ctx, idGenerado)
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(nuevaImagen)
}

// GET /imagenes/{id}
func getImagen(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()
	imagen, err := queries.GetImagen(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Imagen no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(imagen)
}

// PUT /imagenes/{id}
func updateImagen(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	var params db.UpdateImagenParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	if strings.TrimSpace(params.Url) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "La URL es obligatoria"})
		return
	}

	params.IDImagen = id

	if _, err := queries.GetImagen(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Imagen no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.UpdateImagen(ctx, params); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al actualizar la imagen"})
		return
	}

	imagenActualizada, _ := queries.GetImagen(ctx, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(imagenActualizada)
}

// DELETE /imagenes/{id}
func deleteImagen(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	if _, err := queries.GetImagen(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Imagen no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.DeleteImagen(ctx, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al eliminar"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
