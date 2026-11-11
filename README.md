# Trabajo Práctico 5: TuAlquiler

**TuAlquiler** es una app de alquileres: cualquiera puede ver las propiedades publicadas, y quien tiene una cuenta puede publicar las suyas con fotos.

Integrantes: Valentin Tonelotto y Clara Yunis

## Requisitos

Se debe tener instalado previamente:
* Go 1.26 o superior
* Docker (con Docker Compose)
* sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
* templ (`go install github.com/a-h/templ/cmd/templ@latest`)
* Hurl
* make y curl

## Comandos

Todos se ejecutan desde la raíz del proyecto.

| Comando | Qué hace |
|---|---|
| `make run` | Levanta la app en `http://localhost:8080/` |
| `make test` | Corre todas las pruebas y deja todo limpio |
| `make db-down` | Baja la base de datos y borra los datos y las fotos |
| `make clean` | Borra el programa compilado |

### `make run`: levantar la app

Genera el código (sqlc y templ), compila, levanta la base de datos con Docker y arranca el servidor. La terminal queda ocupada mientras el servidor está corriendo.

Después hay que abrir `http://localhost:8080/` en el navegador. Para terminar, se detiene el servidor con Ctrl+C y se baja la base con `make db-down`.

### `make test`: correr las pruebas

Hace todo en un solo comando, en este orden:

1. Revisa que estén instaladas las herramientas y que el puerto 8080 esté libre.
2. Genera el código (sqlc y templ) y compila.
3. Levanta una base de datos vacía con Docker y espera a que esté lista.
4. Arranca el servidor y espera a que responda.
5. Corre las pruebas del archivo `requests.hurl`.
6. Limpia: apaga el servidor, baja la base y borra las fotos de prueba.

Las pruebas cubren crear una cuenta, ingresar y salir, publicar con fotos, ver cada publicación, los errores de los formularios y que nadie pueda borrar publicaciones de otra persona.

Para correrlo, la app no tiene que estar levantada (si el puerto 8080 está ocupado, avisa y no hace nada) y **se borran los datos que haya en la base**. El detalle de cada paso está en `scripts/test.sh`.

## Cómo probarla en el navegador

La base arranca sin usuarios ni publicaciones. Con la app levantada (`make run`):

1. Entrar a `http://localhost:8080/` y tocar **Ingresar** → **Crear cuenta**.
2. Tocar **+ Publicar propiedad**, completar los datos, elegir algunas fotos y publicar.
3. La publicación aparece en la página principal. Tocándola se abre su página, con las fotos y los datos de contacto.
4. Con **Cerrar sesión** se ve la página como la vería alguien sin cuenta.

## Qué se puede hacer

- **Ver todas las publicaciones** en la página principal, sin necesidad de tener cuenta.
- **Ver cada publicación** en su propia página, con todas sus fotos, la descripción y los datos para contactar al dueño (email y teléfono).
- **Crear una cuenta, ingresar y salir.**
- **Publicar una propiedad** con fotos (hasta 8, de 5 MB como máximo cada una). Para publicar hay que tener la sesión iniciada.
- **Eliminar** las publicaciones propias. Se borran directamente, junto con sus fotos.

## Cómo está organizado

```
main.go          → arranca el servidor y registra las rutas
handlers/        → reciben cada pedido, consultan la base y eligen qué página mostrar
views/           → las páginas, escritas con templ (.templ)
static/          → la hoja de estilos y el logo
db/              → el esquema, las consultas (sqlc) y las ciudades iniciales
scripts/test.sh  → el script que usa make test
requests.hurl    → las pruebas
uploads/         → las fotos que suben los usuarios (se crea sola)
```

Las páginas están divididas en componentes:
- `layout.templ`: la estructura que comparten todas las páginas (encabezado con el logo y el botón para ingresar o salir).
- `entity_list.templ`: la lista de publicaciones.
- `entity_form.templ`: el formulario para publicar.
- `index.templ`: la página principal, que junta el formulario y la lista.
- `publicacion.templ`: la página de cada publicación.
- `login.templ`: las páginas para ingresar y crear una cuenta.

### Sesión

Al ingresar, el servidor guarda una cookie con un token que dura 5 minutos. Mientras se use la página, el token se renueva solo; si se deja de usar 5 minutos, hay que volver a ingresar. Las contraseñas no se guardan tal cual: se guardan protegidas con bcrypt.

## Base de datos

La base de datos (PostgreSQL) está estructurada de la siguiente manera:
- TABLA USERS: son las personas que interactúan con la plataforma. Se guarda: id_usuario, nombre, apellido, email, password, telefono, dni, fecha de registro.
- TABLA PUBLICACION: representa la propiedad publicada. Se guarda: id_alquiler, titulo, descripcion, id_usuario (relación con la tabla users), direccion_calle, numero_calle, id_ciudad (relación con la tabla ciudad).
- TABLA IMAGEN: id_imagen, id_alquiler (relación con la tabla publicacion), url.
- TABLA CIUDAD: id_ciudad, nombre, provincia.
- TABLA FAVORITOS: id_favorito, id_usuario (relación con la tabla users), id_alquiler (relación con la tabla publicacion).

Restricciones:
- ON DELETE CASCADE: si se elimina un usuario, se borran sus publicaciones y favoritos. Si se elimina una publicación, se borran sus imágenes y favoritos.
- No pueden existir dos usuarios con el mismo DNI o el mismo email.
- Un usuario no puede agregar a favoritos la misma publicación más de una vez.

Al crear la base se cargan 10 ciudades (`db/seed.sql`), que son las que se pueden elegir al publicar.
