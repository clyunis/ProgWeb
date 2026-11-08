// Funciones de autenticación compartidas por todas las páginas.
//
// El token vive en una cookie HttpOnly: JavaScript NO puede leerla ni borrarla.
// El navegador la adjunta solo a cada petición; por eso acá nunca tocamos el token,
// solo preguntamos al backend si la sesión es válida.

const PAGINA_LOGIN = '/login.html';
const MARGEN_REFRESCO_MS = 20 * 1000; // pedimos token nuevo 20 s antes de que venza

let temporizadorRefresco = null;

// URL del login que, después de iniciar sesión, vuelve a la página actual
function urlLogin() {
    return `${PAGINA_LOGIN}?volver=${encodeURIComponent(location.pathname)}`;
}

// Wrapper de fetch: manda la cookie y centraliza el manejo de 401.
// Si el backend responde 401 (token inválido o expirado), vamos al login.
async function fetchApi(url, opciones = {}) {
    const respuesta = await fetch(url, {
        credentials: 'include', // adjuntar la cookie (en el mismo origen ya es el default)
        ...opciones,
    });

    if (respuesta.status === 401) {
        console.log('Sesión expirada, redirigiendo al login.');
        window.location.href = urlLogin();
        throw new Error('No autorizado');
    }
    return respuesta;
}

// Pregunta al backend si hay una sesión activa.
// Devuelve { usuario, expira } o null si no hay sesión.
async function checkAuthStatus() {
    try {
        const respuesta = await fetch('/welcome', { credentials: 'include' });
        if (!respuesta.ok) return null;
        return await respuesta.json();
    } catch {
        return null;
    }
}

// Programa el refresco del token para justo antes de que expire.
// El backend solo acepta /refresh en los últimos 30 segundos de vida del token,
// y nos informa la fecha de expiración en /signin, /welcome y /refresh.
function programarRefresco(expira) {
    clearTimeout(temporizadorRefresco);
    const faltan = new Date(expira).getTime() - Date.now() - MARGEN_REFRESCO_MS;
    temporizadorRefresco = setTimeout(refrescarToken, Math.max(faltan, 0));
}

async function refrescarToken() {
    try {
        const respuesta = await fetchApi('/refresh', { method: 'POST' });
        if (!respuesta.ok) throw new Error(`Error ${respuesta.status}`);
        const datos = await respuesta.json();
        programarRefresco(datos.expira); // y así sucesivamente mientras la página esté abierta
    } catch (err) {
        console.error('No se pudo refrescar la sesión:', err);
        window.location.href = urlLogin();
    }
}

// Logout: el frontend no puede borrar la cookie HttpOnly,
// así que le pide al backend que la invalide.
async function logout() {
    clearTimeout(temporizadorRefresco);
    const respuesta = await fetch('/logout', { method: 'POST', credentials: 'include' });
    if (respuesta.ok) {
        console.log('Sesión cerrada.');
        window.location.href = '/'; // la página principal es pública: volvemos a las publicaciones
    } else {
        console.error('Falló el logout');
    }
}
