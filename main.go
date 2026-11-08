package main

import (
	"database/sql"
	"log"
	"net/http"

	db "tp4/db/sqlc"

	_ "github.com/lib/pq"
)

var queries *db.Queries

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/app_alquileres?sslmode=disable"
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("No se pudo establecer conexión: %v", err)
	}

	queries = db.New(conn)

	// Autenticación con JWT en cookie HttpOnly
	http.HandleFunc("/signin", signinHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/refresh", refreshHandler)
	http.HandleFunc("/logout", logoutHandler)

	// API REST: los GET son públicos; POST, PUT y DELETE requieren sesión
	// (salvo POST /users, que es el registro).
	http.HandleFunc("/publicaciones", conSesionParaEscribir(publicacionesHandler))
	http.HandleFunc("/publicaciones/", conSesionParaEscribir(publicacionesHandler))

	http.HandleFunc("/ciudades", conSesionParaEscribir(ciudadesHandler))
	http.HandleFunc("/ciudades/", conSesionParaEscribir(ciudadesHandler))

	http.HandleFunc("/users", conSesionParaEscribir(usersHandler))
	http.HandleFunc("/users/", conSesionParaEscribir(usersHandler))

	http.HandleFunc("/imagenes", conSesionParaEscribir(imagenesHandler))
	http.HandleFunc("/imagenes/", conSesionParaEscribir(imagenesHandler))

	http.HandleFunc("/favoritos", conSesionParaEscribir(favoritosHandler))
	http.HandleFunc("/favoritos/", conSesionParaEscribir(favoritosHandler))

	// Imágenes subidas por los usuarios
	http.Handle("/uploads/", uploadsHandler())

	// Página de detalle: /propiedad/{id} siempre devuelve el mismo HTML,
	// y publicacion.js lee el id de la URL y pide los datos a la API.
	http.HandleFunc("/propiedad/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/publicacion.html")
	})

	// Frontend: sirve index.html, app.js y styles.css desde la carpeta static.
	// Como "/" es la ruta menos específica, solo atiende lo que no matchea con la API.
	http.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("Servidor corriendo en http://localhost:8080 ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
