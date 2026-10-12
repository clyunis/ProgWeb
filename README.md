# Trabajo Práctico 3

En este TP implementamos los handlers para las operaciones CRUD sobre las tablas de nuestro dominio, exponiendo la capa de datos generada con sqlc a través de una API REST con endpoints HTTP.

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
| Publicaciones | `GET /publicaciones`, `POST /publicaciones`, `GET /publicaciones/{id}`, `PUT /publicaciones/{id}`, `DELETE /publicaciones/{id}` |
| Imagenes | `GET /imagenes?id_alquiler={id}`, `POST /imagenes`, `GET /imagenes/{id}`, `PUT /imagenes/{id}`, `DELETE /imagenes/{id}` |
| Favoritos | `GET /favoritos?id_usuario={id}`, `POST /favoritos`, `GET /favoritos/{id}`, `DELETE /favoritos/{id}` |

Las imagenes se listan por publicacion y los favoritos por usuario, por eso esos listados requieren el parametro en la URL. Los favoritos no tienen `PUT` porque un favorito solo se crea o se elimina.

Codigos de respuesta: `200` (OK), `201` (creado), `204` (eliminado), `400` (JSON invalido o datos faltantes), `404` (no existe), `405` (metodo no permitido) y `500` (error de la base de datos).

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

2. En otra terminal, probar la API:

```bash
make hurl
```

Este comando ejecuta todas las peticiones del archivo `requests.hurl` (crear, listar, ver, actualizar y eliminar cada entidad) y verifica que las respuestas sean las esperadas.

3. Para terminar, detener el servidor y bajar la base de datos:

```bash
make db-down
```

Este comando detiene el contenedor y borra los datos de la base. Con `make clean` se borra el binario compilado.
