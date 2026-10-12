package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	db "tp3/db/sqlc"
)

func ciudadesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/ciudades")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			getCiudades(w, r)
		case http.MethodPost:
			createCiudad(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de ciudad inválido"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getCiudad(w, r, int32(id))
	case http.MethodPut:
		updateCiudad(w, r, int32(id))
	case http.MethodDelete:
		deleteCiudad(w, r, int32(id))
	default:
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
	}
}

// GET /ciudades
func getCiudades(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ciudades, err := queries.ListCiudades(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al consultar la base de datos"})
		return
	}

	if ciudades == nil {
		ciudades = []db.Ciudad{} // Retorna array vacío en lugar de null si no hay registros
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ciudades)
}

// POST /ciudades
func createCiudad(w http.ResponseWriter, r *http.Request) {
	var params db.CreateCiudadParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	// Validación simple
	if strings.TrimSpace(params.Nombre) == "" || strings.TrimSpace(params.Provincia) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El nombre y la provincia son obligatorios"})
		return
	}

	ctx := r.Context()
	idGenerado, err := queries.CreateCiudad(ctx, params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al guardar en la base de datos"})
		return
	}

	// Recuperar y devolver la entidad recién creada
	nuevaCiudad, _ := queries.GetCiudad(ctx, idGenerado)
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(nuevaCiudad)
}

// GET /ciudades/{id}
func getCiudad(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()
	ciudad, err := queries.GetCiudad(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ciudad no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ciudad)
}

// PUT /ciudades/{id}
func updateCiudad(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	var params db.UpdateCiudadParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	if strings.TrimSpace(params.Nombre) == "" || strings.TrimSpace(params.Provincia) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El nombre y la provincia son obligatorios"})
		return
	}

	params.IDCiudad = id

	// Verificamos que exista antes de actualizar, para poder devolver 404
	if _, err := queries.GetCiudad(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ciudad no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.UpdateCiudad(ctx, params); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al actualizar la ciudad"})
		return
	}

	ciudadActualizada, _ := queries.GetCiudad(ctx, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ciudadActualizada)
}

// DELETE /ciudades/{id}
func deleteCiudad(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	// Verificamos que exista antes de borrar, para poder devolver 404
	if _, err := queries.GetCiudad(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ciudad no encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.DeleteCiudad(ctx, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al eliminar"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
