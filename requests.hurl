# Tests de TuAlquiler (TP5: páginas renderizadas en el servidor con templ).
# Se corren con "make test", que arranca todo sobre una base vacía.
# Hurl guarda las cookies entre pedidos, como un navegador: después de
# ingresar, los pedidos siguientes viajan con la cookie "token".
# Hurl no sigue las redirecciones, así podemos comprobar el 303 de cada formulario.


# ==========================================================
# PÁGINA PRINCIPAL SIN SESIÓN
# ==========================================================

# 1. La página llega completa desde el servidor, sin JavaScript
GET http://localhost:8080/
HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "TuAlquiler"
body contains "Todavía no hay publicaciones"
body contains "Ingresá y publicalo"
body not contains "action=\"/publicaciones\""
body not contains "<script"

# 2. Archivos estáticos
GET http://localhost:8080/static/styles.css
HTTP 200

# 3. Sin sesión no se puede publicar: redirige al login
POST http://localhost:8080/publicaciones
[MultipartFormData]
titulo: Algo
HTTP 303
[Asserts]
header "Location" == "/login"


# ==========================================================
# REGISTRO
# ==========================================================

# 1. Formulario de registro
GET http://localhost:8080/registro
HTTP 200
[Asserts]
body contains "action=\"/registro\""

# 2. Contraseña demasiado corta: vuelve a mostrar el formulario con el error
POST http://localhost:8080/registro
[FormParams]
nombre: Ana
apellido: Gómez
dni: 30123999
telefono: 11-5555-1234
email: ana.prueba@example.com
password: 123
HTTP 422
[Asserts]
body contains "al menos 6 caracteres"
body contains "value=\"ana.prueba@example.com\""

# 3. Registro correcto: crea la cuenta, pone la cookie y redirige (Post/Redirect/Get)
POST http://localhost:8080/registro
[FormParams]
nombre: Ana
apellido: Gómez
dni: 30123999
telefono: 11-5555-1234
email: ana.prueba@example.com
password: unaClaveSegura123
HTTP 303
[Asserts]
header "Location" == "/"
cookie "token" exists
cookie "token[HttpOnly]" exists
cookie "token[SameSite]" == "Lax"

# 4. Con sesión, la página principal saluda y muestra el formulario para publicar
GET http://localhost:8080/
HTTP 200
[Asserts]
body contains "Hola, Ana"
body contains "action=\"/publicaciones\""
body contains "Tandil, Buenos Aires"


# ==========================================================
# PUBLICAR (formulario con fotos)
# ==========================================================

# 1. Falta el título: vuelve a mostrar el formulario con lo que se había escrito
POST http://localhost:8080/publicaciones
[MultipartFormData]
titulo:
descripcion: 2 ambientes con balcón
direccion_calle: Rivadavia
numero_calle: 1234
id_ciudad: 1
HTTP 422
[Asserts]
body contains "Completá todos los campos."
body contains "value=\"Rivadavia\""

# 2. Una foto que no es una imagen: no se crea nada
POST http://localhost:8080/publicaciones
[MultipartFormData]
titulo: No debería quedar
descripcion: x
direccion_calle: x
numero_calle: 1
id_ciudad: 1
imagenes: file,requests.hurl; image/png
HTTP 422
[Asserts]
body contains "no es una imagen válida"

# 3. Publicación correcta con una foto: redirige a la página principal
POST http://localhost:8080/publicaciones
[MultipartFormData]
titulo: Departamento céntrico
descripcion: 2 ambientes con balcón
direccion_calle: Av. San Martin
numero_calle: 1234
id_ciudad: 1
imagenes: file,testdata/foto.png; image/png
HTTP 303
[Asserts]
header "Location" == "/"

# 4. La publicación aparece en la lista (y la que falló no)
GET http://localhost:8080/
HTTP 200
[Captures]
id_pub: regex /href="\/propiedad\/(\d+)"/
foto_url: regex /src="(\/uploads\/[0-9a-f]+\.png)"/
[Asserts]
body contains "Departamento céntrico"
body contains "Publicado por vos"
body contains "1 propiedad"
body not contains "No debería quedar"

# 5. La foto subida se puede ver
GET http://localhost:8080{{foto_url}}
HTTP 200
[Asserts]
header "Content-Type" == "image/png"

# 6. No se puede listar la carpeta de fotos
GET http://localhost:8080/uploads/
HTTP 404


# ==========================================================
# PÁGINA DE CADA PUBLICACIÓN
# ==========================================================

# 1. Detalle de la publicación (la dueña ve el botón para eliminarla)
GET http://localhost:8080/propiedad/{{id_pub}}
HTTP 200
[Asserts]
body contains "Departamento céntrico"
body contains "{{foto_url}}"
body contains "Esta es tu publicación."
body contains "/publicaciones/{{id_pub}}/eliminar"

# 2. Publicación que no existe
GET http://localhost:8080/propiedad/999999
HTTP 404
[Asserts]
body contains "Publicación no encontrada"


# ==========================================================
# CERRAR SESIÓN E INGRESAR
# ==========================================================

# 1. Cerrar sesión borra la cookie
POST http://localhost:8080/logout
HTTP 303
[Asserts]
header "Location" == "/"
header "Set-Cookie" contains "token=;"

# 2. Ya no hay sesión
GET http://localhost:8080/
HTTP 200
[Asserts]
body contains "Ingresar"
body not contains "Hola, Ana"

# 3. Sin sesión, el detalle muestra igual los datos de contacto
GET http://localhost:8080/propiedad/{{id_pub}}
HTTP 200
[Asserts]
body contains "mailto:ana.prueba@example.com"
body not contains "/eliminar"

# 4. Contraseña incorrecta
POST http://localhost:8080/login
[FormParams]
email: ana.prueba@example.com
password: claveIncorrecta
HTTP 401
[Asserts]
body contains "Email o contraseña incorrectos."
cookie "token" not exists


# ==========================================================
# PERMISOS (un usuario no puede borrar lo que es de otro)
# ==========================================================

# 1. Se registra un segundo usuario
POST http://localhost:8080/registro
[FormParams]
nombre: Beto
apellido: López
dni: 31222333
telefono: 11-5555-0000
email: beto.prueba@example.com
password: otraClave456
HTTP 303

# 2. No ve el botón para eliminar la publicación de Ana
GET http://localhost:8080/propiedad/{{id_pub}}
HTTP 200
[Asserts]
body contains "Hola, Beto"
body not contains "/eliminar"

# 3. Aunque mande el formulario a mano, el servidor no lo deja
POST http://localhost:8080/publicaciones/{{id_pub}}/eliminar
HTTP 403

# 4. No se puede registrar otra cuenta con el mismo email
POST http://localhost:8080/registro
[FormParams]
nombre: Otra
apellido: Ana
dni: 39999999
telefono: 1
email: ana.prueba@example.com
password: 123456
HTTP 409
[Asserts]
body contains "Ya hay una cuenta con ese email o DNI."

# 5. Beto cierra sesión
POST http://localhost:8080/logout
HTTP 303


# ==========================================================
# ELIMINAR
# ==========================================================

# 1. Ana vuelve a ingresar
POST http://localhost:8080/login
[FormParams]
email: ana.prueba@example.com
password: unaClaveSegura123
HTTP 303
[Asserts]
header "Location" == "/"
cookie "token" exists

# 2. Eliminar solo funciona con POST
GET http://localhost:8080/publicaciones/{{id_pub}}/eliminar
HTTP 405

# 3. Ana elimina su publicación
POST http://localhost:8080/publicaciones/{{id_pub}}/eliminar
HTTP 303
[Asserts]
header "Location" == "/"

# 4. Ya no existe la publicación
GET http://localhost:8080/propiedad/{{id_pub}}
HTTP 404

# 5. Y su foto también se borró del disco
GET http://localhost:8080{{foto_url}}
HTTP 404
