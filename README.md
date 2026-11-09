# Trabajo Práctico 5

**TuAlquiler** es una app de alquileres: cualquiera puede ver las propiedades publicadas, y quien tiene una cuenta puede publicar las suyas con fotos.

En el TP4 la página se armaba en el navegador con JavaScript, pidiéndole los datos (en JSON) al servidor. En este TP cambiamos eso: ahora **el servidor arma la página completa** con [templ](https://templ.guide/) y el navegador solo la muestra. La app no usa JavaScript.

Integrantes: Valentin Tonelotto y Clara Yunis

## Base de datos

La base de datos está estructurada de la siguiente manera:
- TABLA USUARIO: son las personas que interactúan con la plataforma. Se guarda: id_usuario, nombre, apellido, email, password, telefono, dni, fecha de registro.
- TABLA PUBLICACION: representa la propiedad publicada. Se guarda: id_alquiler, titulo, descripcion, id_usuario (relación con tabla usuario), direccion_calle, numero_calle, id_ciudad (relación con tabla ciudad).
- TABLA IMAGEN: id_imagen, id_alquiler (relación con la tabla publicacion), url.
- TABLA CIUDAD: id_ciudad, nombre, provincia.
- TABLA FAVORITOS: id_favorito, id_usuario (relación con tabla usuario), id_alquiler (relación con tabla publicacion).

Restricciones:
- ON DELETE CASCADE: si se elimina un usuario, se borran sus publicaciones y favoritos. Si se elimina una publicación, se borran sus imágenes y favoritos.
- No pueden existir dos usuarios con el mismo DNI o el mismo email.
- Un usuario no puede agregar a favoritos la misma publicación más de una vez.

Al crear la base se cargan algunas ciudades (`db/seed.sql`) para poder publicar desde el primer momento.

## Qué se puede hacer

- **Ver todas las publicaciones** en la página principal, sin necesidad de tener cuenta.
- **Ver cada publicación** en su propia página, con todas sus fotos, la descripción y los datos para contactar al dueño.
- **Crear una cuenta, ingresar y salir.**
- **Publicar una propiedad** con fotos (hasta 8, de 5 MB como máximo cada una).
- **Eliminar** las publicaciones propias.

## Cómo está organizado

```
main.go          → arranca el servidor y registra las rutas
handlers/        → reciben cada pedido, consultan la base y eligen qué página mostrar
views/           → las páginas, escritas con templ (.templ)
static/          → la hoja de estilos y el logo
db/              → el esquema, las consultas (sqlc) y las ciudades iniciales
uploads/         → las fotos que suben los usuarios (se crea sola)
```

Las páginas están divididas en componentes:
- `layout.templ`: la estructura que comparten todas las páginas (encabezado con el logo y el botón para ingresar o salir).
- `entity_list.templ`: la lista de publicaciones.
- `entity_form.templ`: el formulario para publicar.
- `index.templ`: la página principal, que junta el formulario y la lista.
- `publicacion.templ`: la página de cada publicación.
- `login.templ`: las páginas para ingresar y crear una cuenta.

### Cómo funcionan los formularios

Todos los formularios (publicar, ingresar, crear cuenta, eliminar) se envían de la forma clásica, sin JavaScript. Cuando el servidor termina de procesar uno, **redirige** a la página principal y el navegador la vuelve a pedir, ya con los cambios. Así, si después se recarga la página, el formulario no se envía dos veces (patrón *Post/Redirect/Get*).

Si algún dato está mal (por ejemplo, falta el título), la página se vuelve a mostrar con un mensaje de error y lo que se había escrito.

### Sesión

Al ingresar, el servidor guarda una cookie con un token que dura 5 minutos. Mientras se use la página, el token se renueva solo; si se deja de usar 5 minutos, hay que volver a ingresar. Las contraseñas se guardan encriptadas.

## Requisitos

Se debe tener instalado previamente:
* Docker
* Go
* sqlc
* templ (`go install github.com/a-h/templ/cmd/templ@latest`)
* Hurl

## Ejecución

Todos los comandos se ejecutan desde la raíz del proyecto.

1. Levantar la app:

```bash
make run
```

Genera el código (sqlc y templ), compila, levanta la base de datos con Docker y arranca el servidor. La terminal queda ocupada mientras el servidor está corriendo.

2. Abrir `http://localhost:8080/` en el navegador.

3. Para terminar, detener el servidor (Ctrl+C) y bajar la base de datos:

```bash
make db-down
```

Esto borra los datos de la base y las fotos subidas. Con `make clean` se borra el programa compilado.

## Tests

```bash
make test
```

Arranca todo sobre una base vacía, ejecuta las pruebas del archivo `requests.hurl` y apaga el servidor. Prueba crear una cuenta, ingresar y salir, publicar con fotos, ver cada publicación, los errores de los formularios y que nadie pueda borrar publicaciones de otra persona.

Para correrlo, el servidor no tiene que estar levantado (usa el puerto 8080) y **se borran los datos que haya en la base**.
