package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"

	db "tp5/db/sqlc"
	"tp5/views"
)

func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		log.Println("Error al renderizar:", err)
	}
}

func (h *Handler) usuarioActual(r *http.Request) *views.Usuario {
	claims, err := validarToken(r)
	if err != nil {
		return nil
	}
	u, err := h.queries.GetUser(r.Context(), claims.IDUsuario)
	if err != nil {
		return nil
	}
	return &views.Usuario{ID: u.IDUsuario, Nombre: u.Nombre, Apellido: u.Apellido}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	h.renderIndex(w, r, h.usuarioActual(r), views.FormPublicacion{}, http.StatusOK)
}

func (h *Handler) renderIndex(w http.ResponseWriter, r *http.Request, usuario *views.Usuario, form views.FormPublicacion, status int) {
	ctx := r.Context()
	pubs, err := h.queries.ListPublicaciones(ctx)
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	imagenes, err := h.queries.ListImagenes(ctx)
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}

	fotos := make(map[int32][]db.Imagen)
	for _, img := range imagenes {
		fotos[img.IDAlquiler] = append(fotos[img.IDAlquiler], img)
	}

	var ciudades []db.Ciudad
	if usuario != nil {
		ciudades, err = h.queries.ListCiudades(ctx)
		if err != nil {
			http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
			return
		}
	}

	render(w, r, status, views.IndexPage(pubs, fotos, ciudades, form, usuario))
}

func (h *Handler) Publicaciones(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/publicaciones")

	if path == "" || path == "/" {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		h.crearPublicacion(w, r)
		return
	}

	idStr := strings.TrimPrefix(path, "/")
	if idPub, ok := strings.CutSuffix(idStr, "/eliminar"); ok {
		id, err := strconv.Atoi(idPub)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		h.eliminarPublicacion(w, r, int32(id))
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) eliminarPublicacion(w http.ResponseWriter, r *http.Request, id int32) {
	usuario := h.usuarioActual(r)
	if usuario == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	ctx := r.Context()

	pub, err := h.queries.GetPublicacion(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		render(w, r, http.StatusNotFound, views.NoEncontradaPage(usuario))
		return
	}
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	if pub.IDUsuario != usuario.ID {
		http.Error(w, "Solo podés eliminar tus propias publicaciones", http.StatusForbidden)
		return
	}

	fotos, err := h.queries.ListImagenesByPublicacion(ctx, id)
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	if err := h.queries.DeletePublicacion(ctx, id); err != nil {
		http.Error(w, "Error al eliminar", http.StatusInternalServerError)
		return
	}
	borrarArchivosSubidos(fotos)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) crearPublicacion(w http.ResponseWriter, r *http.Request) {
	usuario := h.usuarioActual(r)
	if usuario == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxImagenesPorEnvio*maxTamanioImagen+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var errTamanio *http.MaxBytesError
		if errors.As(err, &errTamanio) {
			form := views.FormPublicacion{Error: "Las fotos superan el tamaño máximo permitido."}
			h.renderIndex(w, r, usuario, form, http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	form := views.FormPublicacion{
		Titulo:      strings.TrimSpace(r.FormValue("titulo")),
		Descripcion: strings.TrimSpace(r.FormValue("descripcion")),
		Calle:       strings.TrimSpace(r.FormValue("direccion_calle")),
		Numero:      strings.TrimSpace(r.FormValue("numero_calle")),
	}
	idCiudad, errCiudad := strconv.Atoi(r.FormValue("id_ciudad"))
	form.IDCiudad = int32(idCiudad)
	archivos := r.MultipartForm.File["imagenes"]

	switch {
	case form.Titulo == "" || form.Descripcion == "" || form.Calle == "" || form.Numero == "":
		form.Error = "Completá todos los campos."
	case errCiudad != nil || !h.existeCiudad(r, form.IDCiudad):
		form.Error = "Elegí una ciudad de la lista."
	case len(archivos) > maxImagenesPorEnvio:
		form.Error = "Podés subir hasta 8 fotos."
	}
	if form.Error != "" {
		h.renderIndex(w, r, usuario, form, http.StatusUnprocessableEntity)
		return
	}

	id, err := h.queries.CreatePublicacion(ctx, db.CreatePublicacionParams{
		Titulo:         form.Titulo,
		Descripcion:    form.Descripcion,
		IDUsuario:      usuario.ID,
		DireccionCalle: form.Calle,
		NumeroCalle:    form.Numero,
		IDCiudad:       form.IDCiudad,
	})
	if err != nil {
		log.Println("Error al crear publicación:", err)
		http.Error(w, "Error al guardar en la base de datos", http.StatusInternalServerError)
		return
	}

	if len(archivos) > 0 {
		if err := h.guardarImagenes(ctx, id, archivos); err != nil {

			h.queries.DeletePublicacion(ctx, id)
			form.Error = "No se pudieron guardar las fotos: " + err.Error()
			h.renderIndex(w, r, usuario, form, http.StatusUnprocessableEntity)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) existeCiudad(r *http.Request, id int32) bool {
	_, err := h.queries.GetCiudad(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("Error al consultar ciudad:", err)
	}
	return err == nil
}
