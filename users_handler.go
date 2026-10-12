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

func usersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/users")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			getUsers(w, r)
		case http.MethodPost:
			createUser(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de usuario inválido"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getUser(w, r, int32(id))
	case http.MethodPut:
		updateUser(w, r, int32(id))
	case http.MethodDelete:
		deleteUser(w, r, int32(id))
	default:
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
	}
}

// GET /users
func getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := queries.ListUsers(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al consultar la base de datos"})
		return
	}

	if users == nil {
		users = []db.ListUsersRow{} // Retorna array vacío en lugar de null si no hay registros
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// POST /users
func createUser(w http.ResponseWriter, r *http.Request) {
	var params db.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	// Validación simple
	if strings.TrimSpace(params.Nombre) == "" || strings.TrimSpace(params.Apellido) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El nombre y el apellido son obligatorios"})
		return
	}
	if strings.TrimSpace(params.Dni) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El DNI es obligatorio"})
		return
	}
	if strings.TrimSpace(params.Email) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El email es obligatorio"})
		return
	}
	if strings.TrimSpace(params.Password) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "La contraseña es obligatoria"})
		return
	}

	ctx := r.Context()
	idGenerado, err := queries.CreateUser(ctx, params)
	if err != nil {
		log.Println("Error al crear usuario:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al guardar en la base de datos"})
		return
	}

	nuevoUsuario, _ := queries.GetUser(ctx, idGenerado)
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(nuevoUsuario)
}

// GET /users/{id}
func getUser(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()
	usuario, err := queries.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no encontrado"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usuario)
}

// PUT /users/{id}
func updateUser(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	var params db.UpdateUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el JSON"})
		return
	}

	if strings.TrimSpace(params.Nombre) == "" || strings.TrimSpace(params.Apellido) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El nombre y el apellido son obligatorios"})
		return
	}
	if strings.TrimSpace(params.Dni) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El DNI es obligatorio"})
		return
	}
	if strings.TrimSpace(params.Email) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "El email es obligatorio"})
		return
	}
	if strings.TrimSpace(params.Password) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "La contraseña es obligatoria"})
		return
	}

	params.IDUsuario = id

	// Verificamos que exista antes de actualizar, para poder devolver 404
	if _, err := queries.GetUser(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no encontrado"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.UpdateUser(ctx, params); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al actualizar el usuario"})
		return
	}

	usuarioActualizado, _ := queries.GetUser(ctx, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usuarioActualizado)
}

// DELETE /users/{id}
func deleteUser(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()

	if _, err := queries.GetUser(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no encontrado"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := queries.DeleteUser(ctx, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al eliminar"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
