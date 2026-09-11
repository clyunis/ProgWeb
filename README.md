# Trabajo Práctico 2

En este TP implementamos la capa de persistencia y acceso a datos de la aplicación y sqlc para la generación de código a partir de consultas SQL.
Integrantes: Valentin Tonelotto y Clara Yunis

## Dominio

Nuestra aplicación web consiste en un sistema de visualización, administración y difusión de alquileres de interés turistico. Se podran realizar publicaciones sobre las propiedades a alquilar con su respectiva información (numero de contacto, descripcion, imagenes, etc). El usuario también podra guardar una lista de sus propiedas favoritas.

## Documentacion

La base de datos estará estructurada de la siguiente manera:
- TABLA USUARIO: son las personas que interactuan con la plataforma. Se guarda la informacion: id_usuario, nombre, email, password, telefono, dni, fecha de registro.
- TABLA PUBLICACION: representa la propiedad publicada. Se guarda la informacion: id_alquiler, titulo, descripcion, id_usuario (relacion con tabla usuario), direccion_calle, nro_calle, id_ciudad (relacion con tabla ciudad)
- TABLA IMAGEN: id_imagen, id_alquiler (relacion con la tabla alquiler), url.
- TABLA CIUDAD: id_ciudad, nombre, provincia.
- TABLA FAVORITOS: id_favoritos, id_usuario (relacion con tabla usuario), id_alquiler (relacion con tabla alquiler)

Se configuraron las siguientes restricciones:
- ON DELETE CASCADE: para las tablas usuario y publicacion. Si se elimina un usuario, se borran automaticamente todas sus publicaciones y sus favoritos. Si se elimina una publicacion, se borran automaticamente todas sus imagenes asociadas y los favoritos asociados a ella.
- No pueden existir dos usuarios con el mismo dni o el mismo email.
- Un usuario no puede agregar a favoritos la misma publicacion más de una vez.

## Requisitos

Se debe tener instalado previamente:
* Docker
* Go
* sqlc

## Ejecucion

Para realizar las pruebas, debe posicionarse en la raiz del proyecto y en la terminal ejecutar el comando:

```bash
make test
```

Este comando utiliza el makefile que se encarga de realizar tareas de limpieza previa, levantar el docker, generar el sqlc, conectar con la base de datos, compilar el codigo y ejecutar los tests. Luego limpia los contenedores al finalizar la ejecución.
