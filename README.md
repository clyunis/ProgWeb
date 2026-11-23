# Trabajo Práctico 6

**TuAlquiler** es una app de alquileres: cualquiera puede ver las propiedades publicadas, y quien tiene una cuenta puede publicar las suyas con fotos.

Integrantes: Valentin Tonelotto y Clara Yunis

## Requisitos

Se debe tener instalado previamente:
* Go 1.26 o superior
* Docker (con Docker Compose)
* sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
* templ (`go install github.com/a-h/templ/cmd/templ@latest`)
* Hurl (https://hurl.dev/docs/installation.html)
* make (en Ubuntu/Debian: `sudo apt install make`)
* curl (en Ubuntu/Debian: `sudo apt install curl`)

Además, los puertos **8080** (la app) y **5432** (la base de datos en Docker) tienen que estar libres.

## Comandos

Todos se ejecutan desde la raíz del proyecto.

| Comando | Qué hace |
|---|---|
| `make run` | Levanta la app en `http://localhost:8080/` y, al detenerla, deja todo limpio |
| `make test` | Corre todas las pruebas y deja todo limpio |
| `make db-down` | Baja la base de datos y borra los datos y las fotos (solo hace falta si algo quedó levantado a mano) |
| `make clean` | Borra el programa compilado |

### `make run`: levantar la app

Hace el ciclo completo en un solo comando, en este orden:

1. Revisa que estén instaladas las herramientas y que el puerto 8080 esté libre (si está ocupado, avisa y no toca nada).
2. Limpia lo que haya quedado de una ejecución anterior: baja la base y borra sus datos y las fotos subidas.
3. Genera el código (sqlc y templ) y compila.
4. Levanta una base de datos vacía con Docker y espera a que esté lista.
5. Arranca el servidor. La terminal queda ocupada mientras está corriendo.
6. Al detenerlo con Ctrl+C, o si el servidor o algún paso anterior falla, limpia siempre: apaga el servidor, baja la base y borra las fotos.

Después del paso 5 hay que abrir `http://localhost:8080/` en el navegador. No hace falta correr `make db-down` al terminar.

### `make test`: correr las pruebas

Hace todo en un solo comando, en este orden:

1. Revisa que estén instaladas las herramientas y que el puerto 8080 esté libre.
2. Genera el código (sqlc y templ) y compila.
3. Levanta una base de datos vacía con Docker y espera a que esté lista.
4. Arranca el servidor y espera a que responda.
5. Corre las pruebas del archivo `requests.hurl`.
6. Limpia: apaga el servidor, baja la base y borra las fotos de prueba.

Las pruebas cubren crear una cuenta, ingresar y salir, publicar con fotos, ver cada publicación, pasar las fotos, eliminar publicaciones, los errores de los formularios y que nadie pueda borrar publicaciones de otra persona. También prueban los pedidos que hace HTMX (mandando el header `HX-Request: true`, como lo hace el navegador).

Para correrlo, la app no tiene que estar levantada (si el puerto 8080 está ocupado, avisa y no hace nada) y **se borran los datos que haya en la base**. El detalle de cada paso está en `scripts/test.sh`.

## Cómo probarla en el navegador

La base arranca sin usuarios ni publicaciones. Con la app levantada (`make run`):

1. Entrar a `http://localhost:8080/` y tocar **Ingresar** → **Crear cuenta**.
2. Tocar **+ Publicar propiedad**, completar los datos, elegir algunas fotos y publicar. La publicación aparece en la lista **sin que la página se recargue**, el formulario se cierra vacío y el contador de propiedades se actualiza.
3. Tocando la publicación se abre su página, con las fotos y los datos de contacto. Las fotos se pasan con las flechas o las miniaturas, también sin recargar.
4. En una publicación propia, tocar **Eliminar** y confirmar: la tarjeta desaparece de la lista, también sin recargar.
5. Con **Cerrar sesión** se ve la página como la vería alguien sin cuenta.

## Qué se puede hacer

- **Ver todas las publicaciones** en la página principal, sin necesidad de tener cuenta.
- **Ver cada publicación** en su propia página, con todas sus fotos, la descripción y los datos para contactar al dueño (email y teléfono).
- **Crear una cuenta, ingresar y salir.**
- **Publicar una propiedad** con fotos (hasta 8, de 5 MB como máximo cada una). Para publicar hay que tener la sesión iniciada.
- **Eliminar** las publicaciones propias, desde la lista (con confirmación y sin recargar) o desde su página. Se borran junto con sus fotos.

## Cómo está organizado

```
main.go          → arranca el servidor y registra las rutas
handlers/        → reciben cada pedido, consultan la base y eligen qué página mostrar
views/           → las páginas, escritas con templ (.templ)
static/          → la hoja de estilos y el logo
db/              → el esquema, las consultas (sqlc) y las ciudades iniciales
scripts/run.sh   → el script que usa make run
scripts/test.sh  → el script que usa make test
requests.hurl    → las pruebas
uploads/         → las fotos que suben los usuarios (se crea sola)
```

Las páginas están divididas en componentes:
- `layout.templ`: la estructura que comparten todas las páginas (encabezado con el logo y el botón para ingresar o salir). Carga HTMX en el `<head>`.
- `entity_list.templ`: la lista de publicaciones (`#lista-publicaciones`), el contador y el botón para eliminar cada tarjeta.
- `entity_form.templ`: el formulario para publicar (`#form-publicar`).
- `index.templ`: la página principal, que junta el formulario y la lista.
- `publicacion.templ`: la página de cada publicación.
- `login.templ`: las páginas para ingresar y crear una cuenta.

### HTMX

Con HTMX, estas acciones actualizan solo una parte de la página, sin recargarla:

- **Publicar:** la publicación nueva aparece en la lista, el formulario se vacía y se cierra, y se actualiza el contador. Si falta algún dato, el error aparece en el formulario.
- **Eliminar desde la lista:** después de confirmar, la tarjeta desaparece y se actualiza el contador.
- **Pasar las fotos** de una publicación: cambia solo la galería. La dirección de la página se actualiza, así que el botón Atrás y recargar funcionan igual.


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
