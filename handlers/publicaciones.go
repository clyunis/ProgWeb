package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"

	db "tp6/db/sqlc"
	"tp6/views"
)

func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		log.Println("Error al renderizar:", err)
	}
}

func esHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func redirigir(w http.ResponseWriter, r *http.Request, url string) {
	if esHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *Handler) cargarPublicaciones(ctx context.Context) ([]db.ListPublicacionesRow, map[int32][]db.Imagen, error) {
	pubs, err := h.queries.ListPublicaciones(ctx)
	if err != nil {
		return nil, nil, err
	}
	imagenes, err := h.queries.ListImagenes(ctx)
	if err != nil {
		return nil, nil, err
	}
	fotos := make(map[int32][]db.Imagen)
	for _, img := range imagenes {
		fotos[img.IDAlquiler] = append(fotos[img.IDAlquiler], img)
	}
	return pubs, fotos, nil
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
	usuario := h.usuarioActual(r)
	ctx := r.Context()
	pubs, fotos, err := h.cargarPublicaciones(ctx)
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}

	var ciudades []db.Ciudad
	if usuario != nil {
		ciudades, err = h.queries.ListCiudades(ctx)
		if err != nil {
			http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
			return
		}
	}

	render(w, r, http.StatusOK, views.IndexPage(pubs, fotos, ciudades, views.FormPublicacion{}, usuario))
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

var (
	errNoEncontrada = errors.New("la publicación no existe")
	errNoEsTuya     = errors.New("solo podés eliminar tus propias publicaciones")
)

func (h *Handler) borrarPublicacion(ctx context.Context, usuario *views.Usuario, id int32) error {
	pub, err := h.queries.GetPublicacion(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errNoEncontrada
	}
	if err != nil {
		return err
	}
	if pub.IDUsuario != usuario.ID {
		return errNoEsTuya
	}

	fotos, err := h.queries.ListImagenesByPublicacion(ctx, id)
	if err != nil {
		return err
	}
	if err := h.queries.DeletePublicacion(ctx, id); err != nil {
		return err
	}
	borrarArchivosSubidos(fotos)
	return nil
}

func (h *Handler) eliminarPublicacion(w http.ResponseWriter, r *http.Request, id int32) {
	usuario := h.usuarioActual(r)
	if usuario == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch err := h.borrarPublicacion(r.Context(), usuario, id); {
	case errors.Is(err, errNoEncontrada):
		render(w, r, http.StatusNotFound, views.NoEncontradaPage(usuario))
	case errors.Is(err, errNoEsTuya):
		http.Error(w, "Solo podés eliminar tus propias publicaciones", http.StatusForbidden)
	case err != nil:
		log.Println("Error al eliminar publicación:", err)
		http.Error(w, "Error al eliminar", http.StatusInternalServerError)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (h *Handler) EliminarPublicacion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	usuario := h.usuarioActual(r)
	if usuario == nil {
		redirigir(w, r, "/login")
		return
	}

	switch err := h.borrarPublicacion(r.Context(), usuario, int32(id)); {
	case errors.Is(err, errNoEncontrada):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, errNoEsTuya):
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	case err != nil:
		log.Println("Error al eliminar publicación:", err)
		http.Error(w, "Error al eliminar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "publicacion-eliminada")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Contador(w http.ResponseWriter, r *http.Request) {
	cantidad, err := h.queries.CountPublicaciones(r.Context())
	if err != nil {
		log.Println("Error al contar publicaciones:", err)
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	if cantidad == 0 {
		render(w, r, http.StatusOK, templ.Join(
			views.Contador(0),
			views.PublicacionListVaciaOOB(),
		))
		return
	}
	render(w, r, http.StatusOK, views.Contador(int(cantidad)))
}

func (h *Handler) crearPublicacion(w http.ResponseWriter, r *http.Request) {
	usuario := h.usuarioActual(r)
	if usuario == nil {
		redirigir(w, r, "/login")
		return
	}
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxImagenesPorEnvio*maxTamanioImagen+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var errTamanio *http.MaxBytesError
		if errors.As(err, &errTamanio) {
			form := views.FormPublicacion{Error: "Las fotos superan el tamaño máximo permitido."}
			h.errorFormulario(w, r, form)
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
		h.errorFormulario(w, r, form)
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
			h.errorFormulario(w, r, form)
			return
		}
	}

	pubs, fotos, err := h.cargarPublicaciones(ctx)
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	ciudades, err := h.queries.ListCiudades(ctx)
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, templ.Join(
		views.PublicacionList(pubs, fotos, usuario),
		views.PublicacionFormOOB(ciudades, views.FormPublicacion{}),
		views.ContadorOOB(len(pubs)),
	))
}

func (h *Handler) errorFormulario(w http.ResponseWriter, r *http.Request, form views.FormPublicacion) {
	ciudades, err := h.queries.ListCiudades(r.Context())
	if err != nil {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Retarget", "#form-publicar")
	w.Header().Set("HX-Reswap", "outerHTML")
	render(w, r, http.StatusOK, views.PublicacionForm(ciudades, form))
}

func (h *Handler) existeCiudad(r *http.Request, id int32) bool {
	_, err := h.queries.GetCiudad(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("Error al consultar ciudad:", err)
	}
	return err == nil
}
