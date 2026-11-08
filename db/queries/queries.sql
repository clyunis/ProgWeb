-- TABLA: USER

-- name: GetUser :one
SELECT id_usuario, nombre, apellido, dni, email, telefono, fecha_registro 
FROM USERS
WHERE id_usuario = $1;

-- name: ListUsers :many
SELECT id_usuario, nombre, apellido, dni, email, telefono, fecha_registro 
FROM USERS
ORDER BY apellido, nombre;

-- Usada por el login: es la única query que devuelve el hash de la contraseña.
-- name: GetUserByEmail :one
SELECT id_usuario, nombre, apellido, email, password
FROM USERS
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO USERS (nombre, apellido, dni, email, telefono, password) 
VALUES ($1, $2, $3, $4, $5, $6) 
RETURNING id_usuario;

-- name: UpdateUser :exec
UPDATE USERS
SET nombre = $2, apellido = $3, dni = $4, email = $5, telefono = $6, password = $7 
WHERE id_usuario = $1;

-- name: DeleteUser :exec
DELETE FROM USERS
WHERE id_usuario = $1;

-- TABLA: PUBLICACION

-- name: GetPublicacion :one
SELECT id_alquiler, titulo, descripcion, id_usuario, direccion_calle, numero_calle, id_ciudad 
FROM PUBLICACION 
WHERE id_alquiler = $1;

-- Para la página de detalle: la publicación con el nombre del autor y de la ciudad
-- name: GetPublicacionDetalle :one
SELECT p.id_alquiler, p.titulo, p.descripcion, p.id_usuario, p.direccion_calle, p.numero_calle, p.id_ciudad,
       u.nombre AS autor_nombre, u.apellido AS autor_apellido,
       c.nombre AS ciudad, c.provincia
FROM PUBLICACION p
JOIN USERS u ON u.id_usuario = p.id_usuario
JOIN CIUDAD c ON c.id_ciudad = p.id_ciudad
WHERE p.id_alquiler = $1;

-- Datos de contacto del dueño de una publicación (solo para usuarios con sesión)
-- name: GetContactoPublicacion :one
SELECT u.nombre, u.apellido, u.email, u.telefono
FROM PUBLICACION p
JOIN USERS u ON u.id_usuario = p.id_usuario
WHERE p.id_alquiler = $1;

-- Incluye el nombre del autor y de la ciudad para que la página pública no
-- tenga que pedir GET /users (que expone datos personales) ni GET /ciudades.
-- name: ListPublicaciones :many
SELECT p.id_alquiler, p.titulo, p.descripcion, p.id_usuario, p.direccion_calle, p.numero_calle, p.id_ciudad,
       u.nombre AS autor_nombre, u.apellido AS autor_apellido,
       c.nombre AS ciudad, c.provincia
FROM PUBLICACION p
JOIN USERS u ON u.id_usuario = p.id_usuario
JOIN CIUDAD c ON c.id_ciudad = p.id_ciudad
ORDER BY p.id_alquiler DESC;

-- name: ListPublicacionesByCiudad :many
SELECT id_alquiler, titulo, descripcion, id_usuario, direccion_calle, numero_calle, id_ciudad 
FROM PUBLICACION 
WHERE id_ciudad = $1 
ORDER BY id_alquiler DESC;

-- name: CreatePublicacion :one
INSERT INTO PUBLICACION (titulo, descripcion, id_usuario, direccion_calle, numero_calle, id_ciudad) 
VALUES ($1, $2, $3, $4, $5, $6) 
RETURNING id_alquiler;

-- name: UpdatePublicacion :exec
UPDATE PUBLICACION 
SET titulo = $2, descripcion = $3, id_usuario = $4, direccion_calle = $5, numero_calle = $6, id_ciudad = $7 
WHERE id_alquiler = $1;

-- name: DeletePublicacion :exec
DELETE FROM PUBLICACION 
WHERE id_alquiler = $1;

-- TABLA: IMAGEN

-- name: GetImagen :one
SELECT id_imagen, id_alquiler, url 
FROM IMAGEN 
WHERE id_imagen = $1;

-- name: ListImagenesByPublicacion :many
SELECT id_imagen, id_alquiler, url 
FROM IMAGEN 
WHERE id_alquiler = $1 
ORDER BY id_imagen ASC;

-- Todas las imágenes (la página principal las agrupa por publicación)
-- name: ListImagenes :many
SELECT id_imagen, id_alquiler, url
FROM IMAGEN
ORDER BY id_alquiler, id_imagen;

-- Imágenes de todas las publicaciones de un usuario (para borrar los archivos
-- cuando se elimina la cuenta, ya que el ON DELETE CASCADE solo borra las filas)
-- name: ListImagenesByUsuario :many
SELECT i.id_imagen, i.id_alquiler, i.url
FROM IMAGEN i
JOIN PUBLICACION p ON p.id_alquiler = i.id_alquiler
WHERE p.id_usuario = $1;

-- name: CreateImagen :one
INSERT INTO IMAGEN (id_alquiler, url) 
VALUES ($1, $2) 
RETURNING id_imagen;

-- name: UpdateImagen :exec
UPDATE IMAGEN 
SET id_alquiler = $2, url = $3 
WHERE id_imagen = $1;

-- name: DeleteImagen :exec
DELETE FROM IMAGEN 
WHERE id_imagen = $1;

-- TABLA: CIUDAD

-- name: GetCiudad :one
SELECT id_ciudad, nombre, provincia 
FROM CIUDAD 
WHERE id_ciudad = $1;

-- name: ListCiudades :many
SELECT id_ciudad, nombre, provincia 
FROM CIUDAD 
ORDER BY provincia, nombre;

-- name: CreateCiudad :one
INSERT INTO CIUDAD (nombre, provincia) 
VALUES ($1, $2) 
RETURNING id_ciudad;

-- name: UpdateCiudad :exec
UPDATE CIUDAD 
SET nombre = $2, provincia = $3 
WHERE id_ciudad = $1;

-- name: DeleteCiudad :exec
DELETE FROM CIUDAD 
WHERE id_ciudad = $1;

-- TABLA: FAVORITOS

-- name: GetFavorito :one
SELECT id_favorito, id_usuario, id_alquiler 
FROM FAVORITOS 
WHERE id_favorito = $1;

-- name: ListFavoritosByUsuario :many
SELECT id_favorito, id_usuario, id_alquiler 
FROM FAVORITOS 
WHERE id_usuario = $1 
ORDER BY id_favorito DESC;

-- name: CreateFavorito :one
INSERT INTO FAVORITOS (id_usuario, id_alquiler) 
VALUES ($1, $2) 
RETURNING id_favorito;

-- name: DeleteFavorito :exec
DELETE FROM FAVORITOS 
WHERE id_favorito = $1;

-- name: DeleteFavoritoByPar :exec
DELETE FROM FAVORITOS 
WHERE id_usuario = $1 AND id_alquiler = $2;