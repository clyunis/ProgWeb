package handlers

import (
	"context"
	"crypto/rand"
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

	db "tp5/db/sqlc"
)

const (
	directorioUploads   = "uploads"
	prefijoURLUploads   = "/uploads/"
	maxTamanioImagen    = 5 << 20
	maxImagenesPorEnvio = 8
)

var extensionesPermitidas = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func Uploads() http.Handler {
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

func (h *Handler) guardarImagenes(ctx context.Context, idAlquiler int32, archivos []*multipart.FileHeader) error {
	if len(archivos) > maxImagenesPorEnvio {
		return fmt.Errorf("podés subir hasta %d imágenes", maxImagenesPorEnvio)
	}
	if err := os.MkdirAll(directorioUploads, 0o755); err != nil {
		return errors.New("no se pudo preparar el almacenamiento")
	}

	var guardadas []db.Imagen
	deshacer := func() {
		for _, img := range guardadas {
			h.queries.DeleteImagen(ctx, img.IDImagen)
		}
		borrarArchivosSubidos(guardadas)
	}

	for _, archivo := range archivos {
		url, err := guardarArchivo(archivo)
		if err != nil {
			deshacer()
			return fmt.Errorf("%s: %v", archivo.Filename, err)
		}

		id, err := h.queries.CreateImagen(ctx, db.CreateImagenParams{IDAlquiler: idAlquiler, Url: url})
		if err != nil {
			borrarArchivosSubidos([]db.Imagen{{Url: url}})
			deshacer()
			return errors.New("no se pudo guardar la imagen en la base de datos")
		}
		guardadas = append(guardadas, db.Imagen{IDImagen: id, IDAlquiler: idAlquiler, Url: url})
	}
	return nil
}

func guardarArchivo(fh *multipart.FileHeader) (string, error) {
	if fh.Size > maxTamanioImagen {
		return "", fmt.Errorf("supera los %d MB", maxTamanioImagen>>20)
	}

	origen, err := fh.Open()
	if err != nil {
		return "", errors.New("no se pudo leer el archivo")
	}
	defer origen.Close()

	cabecera := make([]byte, 512)
	n, _ := io.ReadFull(origen, cabecera)
	ext, ok := extensionesPermitidas[http.DetectContentType(cabecera[:n])]
	if !ok {
		return "", errors.New("no es una imagen válida (se aceptan JPG, PNG, WEBP y GIF)")
	}
	if _, err := origen.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("no se pudo leer el archivo")
	}

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

func borrarArchivosSubidos(imagenes []db.Imagen) {
	for _, img := range imagenes {
		if !strings.HasPrefix(img.Url, prefijoURLUploads) {
			continue
		}
		nombre := filepath.Base(img.Url)
		if err := os.Remove(filepath.Join(directorioUploads, nombre)); err != nil && !os.IsNotExist(err) {
			log.Println("No se pudo borrar", nombre, err)
		}
	}
}
