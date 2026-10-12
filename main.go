package main

import (
	"database/sql"
	"log"
	"net/http"

	db "tp3/db/sqlc"

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

	http.HandleFunc("/publicaciones", publicacionesHandler)
	http.HandleFunc("/publicaciones/", publicacionesHandler)

	http.HandleFunc("/ciudades", ciudadesHandler)
	http.HandleFunc("/ciudades/", ciudadesHandler)

	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/users/", usersHandler)

	http.HandleFunc("/imagenes", imagenesHandler)
	http.HandleFunc("/imagenes/", imagenesHandler)

	http.HandleFunc("/favoritos", favoritosHandler)
	http.HandleFunc("/favoritos/", favoritosHandler)

	log.Println("Servidor corriendo en el puerto 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
