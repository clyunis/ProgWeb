package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	db "tp4/db/sqlc"
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

	// POST /publicaciones/{id}/imagenes: subir imágenes a una publicación
	if idPub, ok := strings.CutSuffix(idStr, "/imagenes"); ok {
		id, err := strconv.Atoi(idPub)
		if err != nil {
			escribirError(w, http.StatusBadRequest, "ID de publicación inválido")
			return
		}
		if r.Method != http.MethodPost {
			escribirError(w, http.StatusMethodNotAllowed, "Método no permitido")
			return
		}
		subirImagenes(w, r, int32(id))
		return
	}

	// GET /publicaciones/{id}/contacto: datos del dueño, solo con sesión iniciada
	if idPub, ok := strings.CutSuffix(idStr, "/contacto"); ok {
		id, err := strconv.Atoi(idPub)
		if err != nil {
			escribirError(w, http.StatusBadRequest, "ID de publicación inválido")
			return
		}
		if r.Method != http.MethodGet {
			escribirError(w, http.StatusMethodNotAllowed, "Método no permitido")
			return
		}
		// Es un GET, así que conSesionParaEscribir lo deja pasar: lo protegemos explícitamente
		requireAuth(func(w http.ResponseWriter, r *http.Request) {
			getContacto(w, r, int32(id))
		})(w, r)
		return
	}

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
		pubs = []db.ListPublicacionesRow{} // Retorna array vacío en lugar de null si no hay registros
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

	// El autor es el usuario logueado (sale del token), no lo que venga en el JSON
	params.IDUsuario = usuarioActual(r).IDUsuario

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
// Incluye el nombre del autor y de la ciudad (lo usa la página de detalle).
func getPublicacion(w http.ResponseWriter, r *http.Request, id int32) {
	ctx := r.Context()
	pub, err := queries.GetPublicacionDetalle(ctx, id)
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

// GET /publicaciones/{id}/contacto (requiere sesión)
// El email y el teléfono no se exponen en los endpoints públicos: así un bot que
// recorre la página no puede juntar los datos personales de todos los dueños.
func getContacto(w http.ResponseWriter, r *http.Request, id int32) {
	contacto, err := queries.GetContactoPublicacion(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			escribirError(w, http.StatusNotFound, "Publicación no encontrada")
			return
		}
		escribirError(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	escribirJSON(w, http.StatusOK, contacto)
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

	// Solo el dueño puede modificarla, y no puede cambiar el autor
	if pub.IDUsuario != usuarioActual(r).IDUsuario {
		escribirError(w, http.StatusForbidden, "Solo podés modificar tus propias publicaciones")
		return
	}
	params.IDUsuario = pub.IDUsuario

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

	// Solo el dueño puede eliminarla
	if pub.IDUsuario != usuarioActual(r).IDUsuario {
		escribirError(w, http.StatusForbidden, "Solo podés eliminar tus propias publicaciones")
		return
	}

	// Anotamos sus imágenes antes de borrar: el ON DELETE CASCADE elimina
	// las filas de IMAGEN, pero los archivos en disco los tenemos que borrar nosotros.
	imagenes, _ := queries.ListImagenesByPublicacion(ctx, id)

	if err := queries.DeletePublicacion(ctx, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al eliminar"})
		return
	}
	borrarArchivosSubidos(imagenes)

	w.WriteHeader(http.StatusNoContent) // 204
}
