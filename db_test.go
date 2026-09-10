package main

import (
	"context"
	"database/sql"
	"testing"

	sqlc "tp2/db/sqlc"

	_ "github.com/lib/pq"
)

const connStr = "user=postgres password=postgres dbname=app_alquileres sslmode=disable"

func TestQueriesUsuario(t *testing.T) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal("Error al abrir la base de datos: ", err)
	}
	defer db.Close()

	queries := sqlc.New(db)
	ctx := context.Background()

	var user_id int32

	t.Run("CreateUser", func(t *testing.T) {
		// CreateUser devuelve (int32, error)
		id, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
			Nombre:   "Test",
			Apellido: "User",
			Dni:      "12345678",
			Email:    "test@example.com",
			Telefono: "1234567890",
			Password: "password",
		})
		if err != nil {
			t.Fatal("Error al crear usuario: ", err)
		}
		user_id = id
	})

	t.Run("GetUser", func(t *testing.T) {
		user, err := queries.GetUser(ctx, user_id)
		if err != nil {
			t.Fatal("Error al obtener usuario: ", err)
		}
		if user.Nombre != "Test" || user.Apellido != "User" || user.Dni != "12345678" || user.Email != "test@example.com" {
			t.Fatal("Error: los datos del usuario no coinciden")
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		err := queries.UpdateUser(ctx, sqlc.UpdateUserParams{
			IDUsuario: user_id,
			Nombre:    "Updated",
			Apellido:  "User",
			Dni:       "87654321",
			Email:     "updated@example.com",
			Telefono:  "0987654321",
			Password:  "newpassword",
		})
		if err != nil {
			t.Fatal("Error al actualizar usuario: ", err)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		users, err := queries.ListUsers(ctx)
		if err != nil {
			t.Fatal("Error al listar usuarios: ", err)
		}
		if len(users) == 0 {
			t.Fatal("Error: no se encontraron usuarios")
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		err := queries.DeleteUser(ctx, user_id)
		if err != nil {
			t.Fatal("Error al eliminar usuario: ", err)
		}
	})
}

func TestQueriesCiudad(t *testing.T) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	queries := sqlc.New(db)
	ctx := context.Background()

	var ciudad_id int32

	t.Run("CreateCiudad", func(t *testing.T) {
		id, err := queries.CreateCiudad(ctx, sqlc.CreateCiudadParams{
			Nombre:    "Tandil",
			Provincia: "Buenos Aires",
		})
		if err != nil {
			t.Fatal("Error al crear ciudad: ", err)
		}
		ciudad_id = id
	})

	t.Run("GetCiudad", func(t *testing.T) {
		ciudad, err := queries.GetCiudad(ctx, ciudad_id)
		if err != nil {
			t.Fatal("Error al obtener ciudad: ", err)
		}
		if ciudad.Nombre != "Tandil" {
			t.Fatal("Los datos no coinciden")
		}
	})

	t.Run("UpdateCiudad", func(t *testing.T) {
		err := queries.UpdateCiudad(ctx, sqlc.UpdateCiudadParams{
			IDCiudad:  ciudad_id,
			Nombre:    "Mar del Plata",
			Provincia: "Buenos Aires",
		})
		if err != nil {
			t.Fatal("Error al actualizar ciudad: ", err)
		}
	})

	t.Run("ListCiudades", func(t *testing.T) {
		ciudades, err := queries.ListCiudades(ctx)
		if err != nil || len(ciudades) == 0 {
			t.Fatal("No se encontraron ciudades")
		}
	})

	t.Run("DeleteCiudad", func(t *testing.T) {
		err := queries.DeleteCiudad(ctx, ciudad_id)
		if err != nil {
			t.Fatal("Error al eliminar ciudad: ", err)
		}
	})
}

func TestQueriesPublicacion(t *testing.T) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	queries := sqlc.New(db)
	ctx := context.Background()

	// user_id y ciudad_id ya son enteros devueltos por sqlc
	user_id, _ := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Nombre: "Dueño", Apellido: "Test", Dni: "1111", Email: "due@test.com", Telefono: "123", Password: "pwd",
	})
	ciudad_id, _ := queries.CreateCiudad(ctx, sqlc.CreateCiudadParams{
		Nombre: "Tandil", Provincia: "BA",
	})

	var pub_id int32

	t.Run("CreatePublicacion", func(t *testing.T) {
		id, err := queries.CreatePublicacion(ctx, sqlc.CreatePublicacionParams{
			Titulo:         "Depto Centro",
			Descripcion:    "Luminoso",
			IDUsuario:      user_id,
			DireccionCalle: "Pinto",
			NumeroCalle:    "399",
			IDCiudad:       ciudad_id,
		})
		if err != nil {
			t.Fatal("Error al crear publicacion: ", err)
		}
		pub_id = id
	})

	t.Run("GetPublicacion", func(t *testing.T) {
		pub, err := queries.GetPublicacion(ctx, pub_id)
		if err != nil {
			t.Fatal("Error al obtener publicacion: ", err)
		}
		if pub.Titulo != "Depto Centro" {
			t.Fatal("Los datos no coinciden")
		}
	})

	t.Run("UpdatePublicacion", func(t *testing.T) {
		err := queries.UpdatePublicacion(ctx, sqlc.UpdatePublicacionParams{
			IDAlquiler:     pub_id,
			Titulo:         "Depto Actualizado",
			Descripcion:    "Luminoso",
			IDUsuario:      user_id,
			DireccionCalle: "Pinto",
			NumeroCalle:    "399",
			IDCiudad:       ciudad_id,
		})
		if err != nil {
			t.Fatal("Error al actualizar publicacion: ", err)
		}
	})

	t.Run("ListPublicaciones", func(t *testing.T) {
		pubs, err := queries.ListPublicaciones(ctx)
		if err != nil || len(pubs) == 0 {
			t.Fatal("Error listando publicaciones")
		}
	})

	t.Run("DeletePublicacion", func(t *testing.T) {
		err := queries.DeletePublicacion(ctx, pub_id)
		if err != nil {
			t.Fatal("Error al eliminar publicacion: ", err)
		}

		queries.DeleteUser(ctx, user_id)
		queries.DeleteCiudad(ctx, ciudad_id)
	})
}

func TestQueriesImagenYFavoritos(t *testing.T) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	queries := sqlc.New(db)
	ctx := context.Background()

	user_id, _ := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Nombre: "Inquilino", Apellido: "Test", Dni: "2222", Email: "inq@test.com", Telefono: "123", Password: "pwd",
	})
	ciudad_id, _ := queries.CreateCiudad(ctx, sqlc.CreateCiudadParams{Nombre: "Tandil", Provincia: "BA"})

	pub_id, _ := queries.CreatePublicacion(ctx, sqlc.CreatePublicacionParams{
		Titulo: "Casa Blanca", Descripcion: "Test", IDUsuario: user_id, DireccionCalle: "San Martin", NumeroCalle: "100", IDCiudad: ciudad_id,
	})

	var img_id, fav_id int32

	t.Run("CreateImagen_Y_Favorito", func(t *testing.T) {
		idImg, err := queries.CreateImagen(ctx, sqlc.CreateImagenParams{
			IDAlquiler: pub_id,
			Url:        "http://foto.com/1.jpg",
		})
		if err != nil {
			t.Fatal("Error creando imagen", err)
		}
		img_id = idImg

		idFav, err := queries.CreateFavorito(ctx, sqlc.CreateFavoritoParams{
			IDUsuario:  user_id,
			IDAlquiler: pub_id,
		})
		if err != nil {
			t.Fatal("Error creando favorito", err)
		}
		fav_id = idFav
	})

	t.Run("Get_List_Updates", func(t *testing.T) {
		_, errImg := queries.GetImagen(ctx, img_id)
		_, errFav := queries.GetFavorito(ctx, fav_id)
		if errImg != nil || errFav != nil {
			t.Fatal("Error obteniendo imagen o favorito")
		}

		queries.UpdateImagen(ctx, sqlc.UpdateImagenParams{IDImagen: img_id, IDAlquiler: pub_id, Url: "http://foto.com/2.jpg"})
	})

	t.Run("Delete_All", func(t *testing.T) {
		if err := queries.DeleteFavorito(ctx, fav_id); err != nil {
			t.Fatal("Error borrando favorito")
		}
		if err := queries.DeleteImagen(ctx, img_id); err != nil {
			t.Fatal("Error borrando imagen")
		}

		queries.DeletePublicacion(ctx, pub_id)
		queries.DeleteUser(ctx, user_id)
		queries.DeleteCiudad(ctx, ciudad_id)
	})
}
