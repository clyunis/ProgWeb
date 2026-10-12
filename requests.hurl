# ==========================================================
# USERS
# ==========================================================

# 1. Crear usuario
POST http://localhost:8080/users
Content-Type: application/json

{
    "nombre": "Ana",
    "apellido": "Gómez",
    "dni": "30123999",
    "email": "ana.prueba@example.com",
    "telefono": "11-5555-1234",
    "password": "unaClaveSegura123"
}

HTTP 201
[Asserts]
jsonpath "$.nombre" == "Ana"
jsonpath "$.password" not exists
jsonpath "$.fecha_registro" isIsoDate
[Captures]
id_usuario: jsonpath "$.id_usuario"


# 2. Listar todos los usuarios (GET)
GET http://localhost:8080/users

HTTP 200
[Asserts]
jsonpath "$" count > 0
jsonpath "$[0].password" not exists


# 3. Ver un usuario específico (GET por ID capturado)
GET http://localhost:8080/users/{{id_usuario}}

HTTP 200
[Asserts]
jsonpath "$.email" == "ana.prueba@example.com"


# 4. Actualizar el usuario (PUT)
PUT http://localhost:8080/users/{{id_usuario}}
Content-Type: application/json

{
    "nombre": "Ana",
    "apellido": "Gómez",
    "dni": "30123999",
    "email": "ana.prueba@example.com",
    "telefono": "11-4444-9999",
    "password": "unaClaveSegura123"
}

HTTP 200
[Asserts]
jsonpath "$.telefono" == "11-4444-9999"


# ==========================================================
# CIUDAD
# ==========================================================

# 1. Crear una nueva ciudad (POST)
POST http://localhost:8080/ciudades
Content-Type: application/json

{
    "nombre": "Tandil",
    "provincia": "Buenos Aires"
}

HTTP 201
[Asserts]
jsonpath "$.nombre" == "Tandil"
[Captures]
id_ciudad: jsonpath "$.id_ciudad"


# 2. Listar todas las ciudades (GET)
GET http://localhost:8080/ciudades

HTTP 200
[Asserts]
jsonpath "$" count > 0


# 3. Ver una ciudad específica (GET por ID capturado)
GET http://localhost:8080/ciudades/{{id_ciudad}}

HTTP 200
[Asserts]
jsonpath "$.nombre" == "Tandil"


# 4. Actualizar la ciudad (PUT)
PUT http://localhost:8080/ciudades/{{id_ciudad}}
Content-Type: application/json

{
    "nombre": "Tandil",
    "provincia": "Buenos Aires (actualizada)"
}

HTTP 200
[Asserts]
jsonpath "$.provincia" == "Buenos Aires (actualizada)"


# ==========================================================
# PUBLICACION
# ==========================================================

# 1. Crear una nueva publicación (POST)
POST http://localhost:8080/publicaciones
Content-Type: application/json

{
    "titulo": "Departamento céntrico",
    "descripcion": "2 ambientes con balcón",
    "id_usuario": {{id_usuario}},
    "direccion_calle": "Av. San Martin",
    "numero_calle": "1234",
    "id_ciudad": {{id_ciudad}}
}

HTTP 201
[Asserts]
jsonpath "$.titulo" == "Departamento céntrico"
[Captures]
id_alquiler: jsonpath "$.id_alquiler"


# 2. Listar todas las publicaciones (GET)
GET http://localhost:8080/publicaciones

HTTP 200
[Asserts]
jsonpath "$" count > 0


# 3. Ver una publicación específica (GET por ID capturado)
GET http://localhost:8080/publicaciones/{{id_alquiler}}

HTTP 200
[Asserts]
jsonpath "$.titulo" == "Departamento céntrico"


# 4. Actualizar la publicación (PUT)
PUT http://localhost:8080/publicaciones/{{id_alquiler}}
Content-Type: application/json

{
    "titulo": "Departamento céntrico ACTUALIZADO",
    "descripcion": "2 ambientes con balcón y cochera",
    "id_usuario": {{id_usuario}},
    "direccion_calle": "Av. San Martin",
    "numero_calle": "1234",
    "id_ciudad": {{id_ciudad}}
}

HTTP 200
[Asserts]
jsonpath "$.titulo" == "Departamento céntrico ACTUALIZADO"


# ==========================================================
# IMAGEN
# ==========================================================

# 1. Agregar una imagen a la publicación (POST)
POST http://localhost:8080/imagenes
Content-Type: application/json

{
    "id_alquiler": {{id_alquiler}},
    "url": "https://example.com/foto1.jpg"
}

HTTP 201
[Asserts]
jsonpath "$.url" == "https://example.com/foto1.jpg"
[Captures]
id_imagen: jsonpath "$.id_imagen"


# 2. Listar las imágenes de esa publicación (GET)
GET http://localhost:8080/imagenes?id_alquiler={{id_alquiler}}

HTTP 200
[Asserts]
jsonpath "$" count > 0


# 3. Ver una imagen específica (GET)
GET http://localhost:8080/imagenes/{{id_imagen}}

HTTP 200
[Asserts]
jsonpath "$.url" == "https://example.com/foto1.jpg"


# 4. Actualizar la imagen (PUT)
PUT http://localhost:8080/imagenes/{{id_imagen}}
Content-Type: application/json

{
    "id_alquiler": {{id_alquiler}},
    "url": "https://example.com/foto1-editada.jpg"
}

HTTP 200
[Asserts]
jsonpath "$.url" == "https://example.com/foto1-editada.jpg"


# ==========================================================
# FAVORITOS
# ==========================================================

# 1. Agregar la publicación a favoritos (POST)
POST http://localhost:8080/favoritos
Content-Type: application/json

{
    "id_usuario": {{id_usuario}},
    "id_alquiler": {{id_alquiler}}
}

HTTP 201
[Asserts]
jsonpath "$.id_usuario" == {{id_usuario}}
[Captures]
id_favorito: jsonpath "$.id_favorito"


# 2. Listar los favoritos del usuario (GET)
GET http://localhost:8080/favoritos?id_usuario={{id_usuario}}

HTTP 200
[Asserts]
jsonpath "$" count > 0


# 3. Ver un favorito específico (GET)
GET http://localhost:8080/favoritos/{{id_favorito}}

HTTP 200
[Asserts]
jsonpath "$.id_alquiler" == {{id_alquiler}}


# ==========================================================
# ERRORES
# ==========================================================

# 1. Ver una publicación inexistente
GET http://localhost:8080/publicaciones/999999

HTTP 404


# 2. Crear una publicación sin título
POST http://localhost:8080/publicaciones
Content-Type: application/json

{
    "titulo": "",
    "descripcion": "Sin título",
    "id_usuario": {{id_usuario}},
    "direccion_calle": "Av. San Martin",
    "numero_calle": "1234",
    "id_ciudad": {{id_ciudad}}
}

HTTP 400


# ==========================================================
# DELETE
# ==========================================================

# 1. Eliminar favorito
DELETE http://localhost:8080/favoritos/{{id_favorito}}
HTTP 204

# 2. Eliminar imagen
DELETE http://localhost:8080/imagenes/{{id_imagen}}
HTTP 204

# 3. Eliminar publicación
DELETE http://localhost:8080/publicaciones/{{id_alquiler}}
HTTP 204

# 4. Eliminar ciudad (Debe ir después de la publicación por la Foreign Key)
DELETE http://localhost:8080/ciudades/{{id_ciudad}}
HTTP 204

# 5. Eliminar usuario
DELETE http://localhost:8080/users/{{id_usuario}}
HTTP 204
