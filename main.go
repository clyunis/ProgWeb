package main

import (
	"database/sql"
	"log"
	"net/http"

	db "tp5/db/sqlc"
	"tp5/handlers"

	_ "github.com/lib/pq"
)

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

	// Como en las filminas: sqlc -> queries -> handlers
	queries := db.New(conn)
	h := handlers.New(queries)

	// Archivos estáticos (CSS, logo). StripPrefix quita "/static/" antes de buscar
	// en la carpeta: /static/styles.css -> static/styles.css
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Imágenes subidas por los usuarios
	http.Handle("/uploads/", handlers.Uploads())

	// Páginas renderizadas con templ
	http.HandleFunc("/", h.Index)
	http.HandleFunc("/publicaciones", h.Publicaciones)
	http.HandleFunc("/publicaciones/", h.Publicaciones)
	http.HandleFunc("/propiedad/", h.Propiedad)

	// Sesión: formularios de ingreso, registro y salida
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/registro", h.Registro)
	http.HandleFunc("/logout", h.Logout)

	// Todas las peticiones pasan por SesionDeslizante, que renueva el token si está por vencer
	log.Println("Servidor corriendo en http://localhost:8080 ...")
	log.Fatal(http.ListenAndServe(":8080", handlers.SesionDeslizante(http.DefaultServeMux)))
}
