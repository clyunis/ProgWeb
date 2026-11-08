// Funciones compartidas por la página principal y la página de cada publicación.
// (Necesita auth.js cargado antes.)

// Atajo para crear elementos: crear('p', 'clase', 'texto')
function crear(etiqueta, clase, texto) {
    const el = document.createElement(etiqueta);
    if (clase) el.className = clase;
    if (texto !== undefined) el.textContent = texto; // textContent: nunca se interpreta como HTML
    return el;
}

// Hace la petición con fetchApi (manda la cookie y redirige al login si hay 401)
// y devuelve el JSON. Si la respuesta no es 2xx, lanza un error con el mensaje
// de la API y su código en err.status.
async function pedir(url, opciones = {}) {
    const res = await fetchApi(url, opciones);
    if (!res.ok) {
        let detalle = `Error ${res.status}`;
        try {
            const cuerpo = await res.json();
            if (cuerpo.error) detalle = cuerpo.error;
        } catch {
            // la respuesta no tenía JSON, nos quedamos con el código
        }
        const err = new Error(detalle);
        err.status = res.status;
        throw err;
    }
    // 204 No Content no tiene cuerpo
    return res.status === 204 ? null : res.json();
}

// Barra superior: "Ingresar" sin sesión, o saludo + "Cerrar sesión" con sesión
function dibujarBarraSesion(usuario) {
    const zona = document.getElementById('zona-sesion');
    zona.innerHTML = '';

    if (usuario) {
        const salir = crear('button', 'btn-secundario', 'Cerrar sesión');
        salir.addEventListener('click', logout);
        zona.append(crear('span', 'saludo', `Hola, ${usuario.nombre}`), salir);
    } else {
        const ingresar = crear('a', 'boton', 'Ingresar');
        ingresar.href = urlLogin(); // después del login vuelve a esta misma página
        zona.appendChild(ingresar);
    }
}
