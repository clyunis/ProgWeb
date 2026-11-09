# Trabajo Práctico 4

Nuestra aplicación es una app de alquileres: cualquiera puede ver las propiedades publicadas, y quien inicia sesión puede publicar las suyas con fotos.

En el TP3 implementamos los handlers para las operaciones CRUD sobre las tablas de nuestro dominio, exponiendo la capa de datos generada con sqlc a través de una API REST con endpoints HTTP.

En este TP agregamos una interfaz de usuario (frontend) en HTML, CSS y JavaScript que consume esa API para listar, crear y eliminar publicaciones, junto con autenticación por JWT y subida de imágenes.

Integrantes: Valentin Tonelotto y Clara Yunis

## Documentacion

La base de datos estará estructurada de la siguiente manera:
- TABLA USUARIO: son las personas que interactuan con la plataforma. Se guarda la informacion: id_usuario, nombre, apellido, email, password, telefono, dni, fecha de registro.
- TABLA PUBLICACION: representa la propiedad publicada. Se guarda la informacion: id_alquiler, titulo, descripcion, id_usuario (relacion con tabla usuario), direccion_calle, numero_calle, id_ciudad (relacion con tabla ciudad)
- TABLA IMAGEN: id_imagen, id_alquiler (relacion con la tabla alquiler), url.
- TABLA CIUDAD: id_ciudad, nombre, provincia.
- TABLA FAVORITOS: id_favorito, id_usuario (relacion con tabla usuario), id_alquiler (relacion con tabla alquiler)

Se configuraron las siguientes restricciones:
- ON DELETE CASCADE: para las tablas usuario y publicacion. Si se elimina un usuario, se borran automaticamente todas sus publicaciones y sus favoritos. Si se elimina una publicacion, se borran automaticamente todas sus imagenes asociadas y los favoritos asociados a ella.
- No pueden existir dos usuarios con el mismo dni o el mismo email.
- Un usuario no puede agregar a favoritos la misma publicacion más de una vez.

## Endpoints

Todas las peticiones y respuestas usan JSON. El servidor corre en `http://localhost:8080`.

| Entidad | Endpoints |
|---|---|
| Usuarios | `GET /users`, `POST /users`, `GET /users/{id}`, `PUT /users/{id}`, `DELETE /users/{id}` |
| Ciudades | `GET /ciudades`, `POST /ciudades`, `GET /ciudades/{id}`, `PUT /ciudades/{id}`, `DELETE /ciudades/{id}` |
| Publicaciones | `GET /publicaciones`, `POST /publicaciones`, `GET /publicaciones/{id}`, `PUT /publicaciones/{id}`, `DELETE /publicaciones/{id}`, `GET /publicaciones/{id}/contacto` (requiere sesión) |
| Imagenes | `GET /imagenes`, `GET /imagenes?id_alquiler={id}`, `POST /imagenes`, `GET /imagenes/{id}`, `PUT /imagenes/{id}`, `DELETE /imagenes/{id}`, `POST /publicaciones/{id}/imagenes` (subida de archivos) |
| Favoritos | `GET /favoritos?id_usuario={id}`, `POST /favoritos`, `GET /favoritos/{id}`, `DELETE /favoritos/{id}` |

`GET /publicaciones` y `GET /publicaciones/{id}` incluyen, además de los campos de la tabla, el nombre del autor (`autor_nombre`, `autor_apellido`) y de la ciudad (`ciudad`, `provincia`), así las páginas públicas no necesitan pedir datos personales de los usuarios. El email y el teléfono del dueño solo se obtienen con `GET /publicaciones/{id}/contacto`, que requiere sesión (`401` si no la hay): así un bot no puede juntar los datos de contacto de todos los dueños. `GET /imagenes` sin parámetro devuelve todas las imágenes; con `id_alquiler`, solo las de esa publicación. Los favoritos se listan por usuario. Los favoritos no tienen `PUT` porque un favorito solo se crea o se elimina.

Codigos de respuesta: `200` (OK), `201` (creado), `204` (eliminado), `400` (JSON invalido o datos faltantes), `401` (sin sesión), `403` (sin permiso), `404` (no existe), `405` (metodo no permitido), `413` (archivos demasiado grandes) y `500` (error de la base de datos).

Al crear la base se cargan algunas ciudades iniciales (`db/seed.sql`) para que se pueda publicar desde el primer momento.

## Imágenes

Las fotos se suben con `POST /publicaciones/{id}/imagenes` como `multipart/form-data` (campo `imagenes`, uno o más archivos). Solo el dueño de la publicación puede subirlas.

- Se aceptan JPG, PNG, WEBP y GIF, de hasta 5 MB cada una y hasta 8 por envío. El tipo se detecta leyendo el contenido del archivo, no por su nombre ni por lo que declara el cliente.
- Cada archivo se guarda en la carpeta `uploads/` con un nombre aleatorio, y en la tabla `IMAGEN` se registra su URL (por ejemplo `/uploads/3f9a…c1.jpg`), asociada a la publicación.
- La carpeta se sirve en `/uploads/` con `http.FileServer` + `http.StripPrefix` (sin permitir listar su contenido).
- Si una imagen falla, no se guarda ninguna del envío.
- Al eliminar una imagen, una publicación o un usuario, también se borran los archivos del disco (el `ON DELETE CASCADE` solo borra las filas).

## Frontend

Los archivos del cliente están en la carpeta `static/` y los sirve el mismo servidor Go con `http.FileServer` en la ruta `/`. Así la página y la API comparten origen y no hay problemas de CORS.

- `index.html`: página principal **pública**. Muestra todas las publicaciones en la lista `#lista-publicaciones`, y el formulario para publicar dentro de un `<dialog>`.
- `app.js`:
  - Al cargar la página consulta si hay sesión (`/welcome`). Sin sesión muestra un botón para ingresar; con sesión, el saludo, "Cerrar sesión" y "+ Publicar propiedad".
  - Pide `GET /publicaciones` y `GET /imagenes`, agrupa las imágenes por publicación y genera cada tarjeta con `createElement` y `textContent` (para evitar XSS). Las tarjetas con varias fotos tienen miniaturas para cambiar la portada.
  - El formulario usa `preventDefault()`, hace `POST /publicaciones` con los datos en JSON y, si se eligieron fotos, las sube con `POST /publicaciones/{id}/imagenes` usando `FormData`. Si la subida falla, borra la publicación para no dejarla a medias. Después vuelve a pedir la lista completa.
  - Las fotos elegidas se muestran como vista previa (con `URL.createObjectURL`) y se pueden quitar antes de publicar.
  - Delegación de eventos: un único listener en la grilla atiende los botones "Eliminar" (que solo aparecen en las publicaciones propias) y las miniaturas de todas las tarjetas.
- Cada tarjeta lleva a la página de su publicación (el link del título se estira con CSS para cubrir toda la tarjeta).
- `publicacion.html` / `publicacion.js`: **página de detalle**, en `/propiedad/{id}`. El servidor Go devuelve el mismo HTML para cualquier id, y el JS lee el número de la URL y pide `GET /publicaciones/{id}` y `GET /imagenes?id_alquiler={id}`. Muestra:
  - una galería con todas las fotos (flechas, miniaturas y teclado ← →; click en la foto para verla en tamaño completo),
  - la descripción completa (respetando los saltos de línea) y la ubicación,
  - una tarjeta de contacto con el nombre del dueño. Con sesión iniciada pide `GET /publicaciones/{id}/contacto` y muestra el email y el teléfono como links (`mailto:` y `tel:`); sin sesión ofrece "Ingresá para ver el contacto", que después del login vuelve a la misma publicación. Si la publicación es propia, permite eliminarla.
- `comun.js`: funciones compartidas por las dos páginas (`crear`, `pedir` y la barra de sesión).
- `login.js` acepta `?volver=/ruta` para volver a la página de origen después de iniciar sesión (solo rutas internas, para evitar redirecciones a otros sitios).
- `styles.css`: CSS Grid para la grilla de tarjetas (`repeat(auto-fill, minmax(280px, 1fr))`, que se adapta sola al ancho de la pantalla) y Flexbox en la barra superior, los formularios y las tarjetas.
- `logo.svg`: logo de la app (también se usa como favicon).

## Autenticación (JWT en cookie HttpOnly)

La autenticación es *stateless*: el servidor no guarda sesiones. Al iniciar sesión emite un JWT firmado (HS256, librería `github.com/golang-jwt/jwt/v5`) que viaja en una cookie `token` con `HttpOnly` y `SameSite=Lax`. Al ser `HttpOnly`, JavaScript no puede leer el token, lo que lo protege ante XSS; el navegador lo adjunta solo a cada petición.

| Endpoint | Descripción |
|---|---|
| `POST /signin` | Recibe `{"email", "password"}`. Si son correctos devuelve `200`, los datos del usuario, la fecha de expiración y la cookie con el token (dura 5 minutos). Si no, `401`. |
| `GET /welcome` | Verifica el token y devuelve el usuario logueado, o `401`. El frontend lo usa para saber si hay sesión. |
| `POST /refresh` | Emite un token nuevo, pero solo si al actual le quedan menos de 30 segundos (si no, `400`). |
| `POST /logout` | Pisa la cookie con una vacía y vencida para que el navegador la borre. Devuelve `204`. |

Reglas de acceso:
- Las contraseñas se guardan hasheadas con **bcrypt**, nunca en texto plano.
- Los `GET` de la API son públicos. `POST`, `PUT` y `DELETE` requieren sesión (`401` si no hay token válido), salvo `POST /users`, que es el registro.
- El autor de una publicación es el usuario logueado (sale del token, no del JSON). Solo el dueño puede modificarla o eliminarla (`403` si no).
- Cada usuario solo puede modificar o eliminar su propia cuenta (`403`).
- La clave para firmar los tokens se lee de la variable de entorno `JWT_SECRET` (si no está, usa una clave de desarrollo).

En el frontend:
- `login.html` / `login.js`: formularios de ingreso y registro (al registrarse inicia sesión automáticamente).
- `auth.js`: `fetchApi` (wrapper de `fetch` que redirige al login ante un `401`), `checkAuthStatus` (consulta `/welcome`), el refresco automático del token unos segundos antes de que venza y `logout`.
- La página principal se puede ver sin sesión; publicar requiere iniciarla. Al cerrar sesión se vuelve a la página principal.

## Requisitos

Se debe tener instalado previamente:
* Docker
* Go
* sqlc
* Hurl

## Ejecucion

Todos los comandos se ejecutan desde la raiz del proyecto.

1. Levantar el servidor:

```bash
make run
```

Este comando genera el codigo con sqlc, compila la aplicacion, levanta la base de datos con docker, espera a que este lista y arranca el servidor en `http://localhost:8080`. La terminal queda ocupada mientras el servidor esta corriendo.

2. Abrir el frontend en el navegador: `http://localhost:8080/`. Las publicaciones se ven sin iniciar sesión; para publicar hay que ingresar o crear una cuenta.

3. En otra terminal, probar la API:

```bash
make hurl
```

Este comando ejecuta todas las peticiones del archivo `requests.hurl` (login, logout, accesos sin sesión, permisos entre usuarios, subida de imágenes, y crear, listar, ver, actualizar y eliminar cada entidad) y verifica que las respuestas sean las esperadas.

4. Para terminar, detener el servidor y bajar la base de datos:

```bash
make db-down
```

Este comando detiene el contenedor, borra los datos de la base y las imágenes subidas (`uploads/`). Con `make clean` se borra el binario compilado.
