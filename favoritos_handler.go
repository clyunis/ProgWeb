package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "tp3/db/sqlc"
)

func favoritosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/favoritos")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			getFavoritosByUsuario(w, r)
		case http.MethodPost:
			createFavorito(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de favorito inválido"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getFavorito(w, r, int32(id))
	case http.MethodDelete:
		deleteFavorito(w, r, int32(id))
	default:
		// No se implementa PUT porque un favorito (usuario-alquiler) no se actualiza, se elimina o se crea.
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
	}
}

// GET /favoritos?id_usuario={id}
func getFavoritosByUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_usuario")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El parámetro id_usuario es obligatorio"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id_usuario inválido"})
		return
	}

	ctx := r.Context()
	favoritos, err := queries.ListFavoritosByUsuario(ctx, int32(id))
	if err != nil {
		log.Println("Error al listar favoritos:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al consultar la base de datos"})
		return
	}

	if favoritos == nil {
		favoritos = []db.Favorito{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(favoritos)
}

// POST /favoritos
func createFavorito(w http.ResponseWriter, r *http.Request) {
	var params db.CreateFavoritoParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	if params.IDUsuario <= 0 || params.IDAlquiler <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id_usuario e id_alquiler son obligatorios"})
		return
	}

	ctx := r.Context()
	idGenerado, err := queries.CreateFavorito(ctx, params)
	if err != nil {
		log.Println("Error al crear favorito:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al guardar en la base de datos (posible duplicado o FK inexistente)"})
		return
	}

	nuevoFavorito, _ := queries.GetFavorito(ctx, idGenerado)
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(nuevoFavorito)
}

// GET /favoritos/{id}
func getFavorito(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()
	favorito, err := queries.GetFavorito(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Favorito no encontrado"})
			return
		}
		log.Println("Error al obtener favorito:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(favorito)
}

// DELETE /favoritos/{id}
func deleteFavorito(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	if _, err := queries.GetFavorito(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Favorito no encontrado"})
			return
		}
		log.Println("Error al verificar favorito para delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.DeleteFavorito(ctx, id); err != nil {
		log.Println("Error al eliminar favorito:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al eliminar"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
