# ==========================================================
# USERS
# ==========================================================

# Hurl guarda las cookies entre requests (como un navegador): después de
# POST /signin, todas las peticiones siguientes viajan con la cookie "token".

# 1. Crear usuario (registro: es público, no requiere sesión)
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


# ==========================================================
# AUTENTICACION (JWT en cookie HttpOnly)
# ==========================================================

# 1. Sin cookie no hay sesión
GET http://localhost:8080/welcome

HTTP 401


# 2. Sin sesión no se puede escribir en la API
POST http://localhost:8080/ciudades
Content-Type: application/json

{
    "nombre": "Ciudad sin sesión",
    "provincia": "Ninguna"
}

HTTP 401


# 3. Login con contraseña incorrecta
POST http://localhost:8080/signin
Content-Type: application/json

{
    "email": "ana.prueba@example.com",
    "password": "claveIncorrecta"
}

HTTP 401
[Asserts]
cookie "token" not exists


# 4. Login con un email que no existe (mismo mensaje que contraseña incorrecta)
POST http://localhost:8080/signin
Content-Type: application/json

{
    "email": "nadie@example.com",
    "password": "unaClaveSegura123"
}

HTTP 401
[Asserts]
jsonpath "$.error" == "Email o contraseña incorrectos"


# 5. Login correcto: el servidor devuelve la cookie HttpOnly con el JWT
POST http://localhost:8080/signin
Content-Type: application/json

{
    "email": "ana.prueba@example.com",
    "password": "unaClaveSegura123"
}

HTTP 200
[Asserts]
cookie "token" exists
cookie "token[HttpOnly]" exists
cookie "token[SameSite]" == "Lax"
cookie "token[Path]" == "/"
jsonpath "$.usuario.id_usuario" == {{id_usuario}}
jsonpath "$.usuario.password" not exists
jsonpath "$.expira" isIsoDate


# 6. Con la cookie, /welcome devuelve el usuario logueado
GET http://localhost:8080/welcome

HTTP 200
[Asserts]
jsonpath "$.usuario.email" == "ana.prueba@example.com"


# 7. El token recién creado todavía no está por vencer: no se refresca
POST http://localhost:8080/refresh

HTTP 400


# ==========================================================
# USERS (con sesión)
# ==========================================================

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

# 1. Crear una nueva publicación (POST). No mandamos id_usuario:
#    el autor es el usuario logueado, sale del token.
POST http://localhost:8080/publicaciones
Content-Type: application/json

{
    "titulo": "Departamento céntrico",
    "descripcion": "2 ambientes con balcón",
    "direccion_calle": "Av. San Martin",
    "numero_calle": "1234",
    "id_ciudad": {{id_ciudad}}
}

HTTP 201
[Asserts]
jsonpath "$.titulo" == "Departamento céntrico"
jsonpath "$.id_usuario" == {{id_usuario}}
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
jsonpath "$.autor_nombre" == "Ana"
jsonpath "$.ciudad" == "Tandil"
jsonpath "$.email" not exists


# 3b. Datos de contacto del dueño (con sesión)
GET http://localhost:8080/publicaciones/{{id_alquiler}}/contacto

HTTP 200
[Asserts]
jsonpath "$.email" == "ana.prueba@example.com"
jsonpath "$.telefono" exists
jsonpath "$.dni" not exists
jsonpath "$.password" not exists


# 3c. Contacto de una publicación que no existe
GET http://localhost:8080/publicaciones/999999/contacto

HTTP 404


# 3d. La página de detalle se sirve en /propiedad/{id}
GET http://localhost:8080/propiedad/{{id_alquiler}}

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "publicacion.js"


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


# 4. Subir archivos de imagen a la publicación (multipart/form-data)
POST http://localhost:8080/publicaciones/{{id_alquiler}}/imagenes
[MultipartFormData]
imagenes: file,testdata/foto.png; image/png
imagenes: file,testdata/foto.png; image/png

HTTP 201
[Asserts]
jsonpath "$" count == 2
jsonpath "$[0].id_alquiler" == {{id_alquiler}}
jsonpath "$[0].url" startsWith "/uploads/"
jsonpath "$[0].url" endsWith ".png"
[Captures]
url_subida: jsonpath "$[0].url"


# 5. El archivo subido se puede descargar
GET http://localhost:8080{{url_subida}}

HTTP 200
[Asserts]
header "Content-Type" == "image/png"


# 6. No se puede listar la carpeta de uploads
GET http://localhost:8080/uploads/

HTTP 404


# 7. Un archivo que no es imagen se rechaza (aunque diga image/png)
POST http://localhost:8080/publicaciones/{{id_alquiler}}/imagenes
[MultipartFormData]
imagenes: file,requests.hurl; image/png

HTTP 400


# 8. GET /imagenes sin parámetro lista todas (la usa la página principal)
GET http://localhost:8080/imagenes

HTTP 200
[Asserts]
jsonpath "$" count >= 3


# 9. Actualizar la imagen (PUT)
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
# PERMISOS (un usuario no puede tocar lo que es de otro)
# ==========================================================

# 1. Crear un segundo usuario
POST http://localhost:8080/users
Content-Type: application/json

{
    "nombre": "Beto",
    "apellido": "López",
    "dni": "31222333",
    "email": "beto.prueba@example.com",
    "telefono": "11-5555-0000",
    "password": "otraClave456"
}

HTTP 201
[Captures]
id_usuario2: jsonpath "$.id_usuario"


# 2. Iniciar sesión como el segundo usuario (la cookie nueva reemplaza a la de Ana)
POST http://localhost:8080/signin
Content-Type: application/json

{
    "email": "beto.prueba@example.com",
    "password": "otraClave456"
}

HTTP 200


# 3. No puede eliminar la publicación de Ana
DELETE http://localhost:8080/publicaciones/{{id_alquiler}}
HTTP 403


# 4. Ni modificarla
PUT http://localhost:8080/publicaciones/{{id_alquiler}}
Content-Type: application/json

{
    "titulo": "Me la quedo",
    "descripcion": "x",
    "direccion_calle": "x",
    "numero_calle": "1",
    "id_ciudad": {{id_ciudad}}
}

HTTP 403


# 5. Ni subirle imágenes
POST http://localhost:8080/publicaciones/{{id_alquiler}}/imagenes
[MultipartFormData]
imagenes: file,testdata/foto.png; image/png

HTTP 403


# 6. Ni eliminar la cuenta de Ana
DELETE http://localhost:8080/users/{{id_usuario}}
HTTP 403


# 7. Sí puede eliminar su propia cuenta
DELETE http://localhost:8080/users/{{id_usuario2}}
HTTP 204


# 8. Volvemos a iniciar sesión como Ana para el resto de las pruebas
POST http://localhost:8080/signin
Content-Type: application/json

{
    "email": "ana.prueba@example.com",
    "password": "unaClaveSegura123"
}

HTTP 200


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

# 3b. Sus archivos de imagen también se borraron del disco
GET http://localhost:8080{{url_subida}}
HTTP 404

# 4. Eliminar ciudad (Debe ir después de la publicación por la Foreign Key)
DELETE http://localhost:8080/ciudades/{{id_ciudad}}
HTTP 204


# ==========================================================
# LOGOUT
# ==========================================================

# 1. Cerrar sesión: el servidor pisa la cookie con una vencida
POST http://localhost:8080/logout

HTTP 204
[Asserts]
header "Set-Cookie" contains "token=;"
header "Set-Cookie" contains "Max-Age=0"


# 2. Ya no hay sesión
GET http://localhost:8080/welcome

HTTP 401


# 2b. Sin sesión no se ven los datos de contacto
GET http://localhost:8080/publicaciones/1/contacto

HTTP 401


# 3. Ni se puede escribir
DELETE http://localhost:8080/users/{{id_usuario}}

HTTP 401


# 4. Iniciar sesión otra vez para poder eliminar la cuenta
POST http://localhost:8080/signin
Content-Type: application/json

{
    "email": "ana.prueba@example.com",
    "password": "unaClaveSegura123"
}

HTTP 200


# 5. Eliminar usuario
DELETE http://localhost:8080/users/{{id_usuario}}
HTTP 204


# 6. El token sigue firmado y sin vencer, pero el usuario ya no existe
GET http://localhost:8080/welcome

HTTP 401
