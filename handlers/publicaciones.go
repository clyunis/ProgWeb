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

// render escribe un componente templ en la respuesta con el código de estado indicado.
func render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		log.Println("Error al renderizar:", err)
	}
}

// usuarioActual devuelve el usuario con sesión iniciada, o nil si no hay sesión
// (no hay cookie, el token es inválido o venció, o el usuario ya no existe).
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

// GET / : página principal con todas las publicaciones.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	// "/" atiende todo lo que no matchea otra ruta: solo respondemos la raíz exacta
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

// renderIndex obtiene los datos con sqlc y renderiza la página principal.
// Se usa en GET / y también cuando el formulario de publicar tiene errores.
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

	// Agrupamos las imágenes por publicación (vienen ordenadas, la primera es la portada)
	fotos := make(map[int32][]db.Imagen)
	for _, img := range imagenes {
		fotos[img.IDAlquiler] = append(fotos[img.IDAlquiler], img)
	}

	// Las ciudades solo hacen falta para el formulario, que se ve con sesión iniciada
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

// Publicaciones atiende /publicaciones y /publicaciones/...
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

	// POST /publicaciones/{id}/eliminar
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

// POST /publicaciones/{id}/eliminar : borra la publicación (solo su dueño) y vuelve a "/".
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

	// Anotamos sus fotos antes de borrar: el ON DELETE CASCADE elimina las filas
	// de IMAGEN, pero los archivos en disco los tenemos que borrar nosotros.
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

	// Post/Redirect/Get: volvemos a la lista, ya sin la publicación
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// POST /publicaciones : crea una publicación con sus fotos (formulario multipart).
// Si sale bien redirige a "/" (Post/Redirect/Get). Si hay errores, vuelve a
// mostrar la página con el mensaje y lo que la persona ya había escrito.
func (h *Handler) crearPublicacion(w http.ResponseWriter, r *http.Request) {
	usuario := h.usuarioActual(r)
	if usuario == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	ctx := r.Context()

	// Limitamos el tamaño total del envío antes de leerlo
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

	// El autor es el usuario con sesión iniciada, nunca un dato del formulario
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
			// Si las fotos fallan, borramos la publicación para no dejarla a medias
			h.queries.DeletePublicacion(ctx, id)
			form.Error = "No se pudieron guardar las fotos: " + err.Error()
			h.renderIndex(w, r, usuario, form, http.StatusUnprocessableEntity)
			return
		}
	}

	// Post/Redirect/Get: el navegador vuelve a pedir "/" y ve la lista actualizada.
	// Si después aprieta F5, repite el GET, no el POST.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) existeCiudad(r *http.Request, id int32) bool {
	_, err := h.queries.GetCiudad(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("Error al consultar ciudad:", err)
	}
	return err == nil
}
