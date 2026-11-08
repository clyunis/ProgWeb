package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	db "tp4/db/sqlc"
)

// Las imágenes subidas se guardan en disco, en esta carpeta, y se sirven en /uploads/.
// En la tabla IMAGEN guardamos solo la URL (ej. "/uploads/3f9a...c1.jpg").
const (
	directorioUploads   = "uploads"
	prefijoURLUploads   = "/uploads/"
	maxTamanioImagen    = 5 << 20 // 5 MB por imagen
	maxImagenesPorEnvio = 8
)

// Tipos de imagen aceptados y la extensión con la que se guardan.
// El tipo se detecta leyendo el contenido del archivo, no confiando en su nombre.
var extensionesPermitidas = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// uploadsHandler sirve los archivos subidos (como en la clase de archivos estáticos:
// FileServer + StripPrefix), pero sin listar el contenido de la carpeta.
func uploadsHandler() http.Handler {
	fs := http.StripPrefix(prefijoURLUploads, http.FileServer(http.Dir(directorioUploads)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fs.ServeHTTP(w, r)
	})
}

// POST /publicaciones/{id}/imagenes (multipart/form-data, campo "imagenes", uno o más archivos).
// Solo el dueño de la publicación puede subirle imágenes.
func subirImagenes(w http.ResponseWriter, r *http.Request, idAlquiler int32) {
	ctx := r.Context()

	pub, err := queries.GetPublicacion(ctx, idAlquiler)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "Publicación no encontrada")
			return
		}
		escribirError(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	if pub.IDUsuario != usuarioActual(r).IDUsuario {
		escribirError(w, http.StatusForbidden, "Solo podés agregar imágenes a tus propias publicaciones")
		return
	}

	// Limitamos el tamaño total del cuerpo antes de leerlo
	r.Body = http.MaxBytesReader(w, r.Body, maxImagenesPorEnvio*maxTamanioImagen+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var errTamanio *http.MaxBytesError
		if errors.As(err, &errTamanio) {
			escribirError(w, http.StatusRequestEntityTooLarge, "Las imágenes superan el tamaño máximo permitido")
			return
		}
		escribirError(w, http.StatusBadRequest, "Se esperaba un formulario multipart con imágenes")
		return
	}
	defer r.MultipartForm.RemoveAll()

	archivos := r.MultipartForm.File["imagenes"]
	if len(archivos) == 0 {
		escribirError(w, http.StatusBadRequest, "No se envió ninguna imagen (campo \"imagenes\")")
		return
	}
	if len(archivos) > maxImagenesPorEnvio {
		escribirError(w, http.StatusBadRequest, fmt.Sprintf("Podés subir hasta %d imágenes", maxImagenesPorEnvio))
		return
	}

	if err := os.MkdirAll(directorioUploads, 0o755); err != nil {
		escribirError(w, http.StatusInternalServerError, "No se pudo preparar el almacenamiento")
		return
	}

	// Guardamos todas o ninguna: si una falla, borramos lo que ya se guardó.
	var guardadas []db.Imagen
	deshacer := func() {
		for _, img := range guardadas {
			queries.DeleteImagen(ctx, img.IDImagen)
		}
		borrarArchivosSubidos(guardadas)
	}

	for _, archivo := range archivos {
		url, err := guardarArchivo(archivo)
		if err != nil {
			deshacer()
			escribirError(w, http.StatusBadRequest, fmt.Sprintf("%s: %v", archivo.Filename, err))
			return
		}

		id, err := queries.CreateImagen(ctx, db.CreateImagenParams{IDAlquiler: idAlquiler, Url: url})
		if err != nil {
			borrarArchivosSubidos([]db.Imagen{{Url: url}})
			deshacer()
			escribirError(w, http.StatusInternalServerError, "Error al guardar la imagen en la base de datos")
			return
		}
		guardadas = append(guardadas, db.Imagen{IDImagen: id, IDAlquiler: idAlquiler, Url: url})
	}

	escribirJSON(w, http.StatusCreated, guardadas)
}

// guardarArchivo valida el archivo y lo copia a la carpeta de uploads con un nombre aleatorio.
// Devuelve la URL pública.
func guardarArchivo(fh *multipart.FileHeader) (string, error) {
	if fh.Size > maxTamanioImagen {
		return "", fmt.Errorf("supera los %d MB", maxTamanioImagen>>20)
	}

	origen, err := fh.Open()
	if err != nil {
		return "", errors.New("no se pudo leer el archivo")
	}
	defer origen.Close()

	// Detectamos el tipo real a partir de los primeros bytes
	cabecera := make([]byte, 512)
	n, _ := io.ReadFull(origen, cabecera)
	ext, ok := extensionesPermitidas[http.DetectContentType(cabecera[:n])]
	if !ok {
		return "", errors.New("no es una imagen válida (se aceptan JPG, PNG, WEBP y GIF)")
	}
	if _, err := origen.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("no se pudo leer el archivo")
	}

	// Nombre aleatorio: evita choques y que el usuario elija rutas (ej. "../../main.go")
	aleatorio := make([]byte, 16)
	rand.Read(aleatorio)
	nombre := hex.EncodeToString(aleatorio) + ext

	destino, err := os.Create(filepath.Join(directorioUploads, nombre))
	if err != nil {
		return "", errors.New("no se pudo guardar el archivo")
	}
	defer destino.Close()

	if _, err := io.Copy(destino, origen); err != nil {
		os.Remove(destino.Name())
		return "", errors.New("no se pudo guardar el archivo")
	}
	return prefijoURLUploads + nombre, nil
}

// borrarArchivosSubidos elimina del disco los archivos de imágenes subidas.
// Las imágenes cargadas como URL externa (ej. https://...) se ignoran.
func borrarArchivosSubidos(imagenes []db.Imagen) {
	for _, img := range imagenes {
		if !strings.HasPrefix(img.Url, prefijoURLUploads) {
			continue
		}
		nombre := filepath.Base(img.Url) // solo el nombre, nunca una ruta
		if err := os.Remove(filepath.Join(directorioUploads, nombre)); err != nil && !os.IsNotExist(err) {
			log.Println("No se pudo borrar", nombre, err)
		}
	}
}
