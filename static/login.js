const formLogin = document.getElementById('form-login');
const formRegistro = document.getElementById('form-registro');
const mensaje = document.getElementById('mensaje');

function mostrarMensaje(texto, tipo) {
    mensaje.textContent = texto;
    mensaje.className = `mensaje ${tipo}`;
}

async function leerError(respuesta) {
    try {
        const cuerpo = await respuesta.json();
        return cuerpo.error || `Error ${respuesta.status}`;
    } catch {
        return `Error ${respuesta.status}`;
    }
}

// A dónde ir después de iniciar sesión: la página desde la que vino (?volver=/propiedad/5)
// o la principal. Solo aceptamos rutas internas ("/algo"), nunca otro sitio
// ("//sitio-malicioso.com" o "https://..."), para que nadie pueda usar
// nuestro login para redirigir a una página falsa.
function destino() {
    const volver = new URLSearchParams(location.search).get('volver');
    if (volver && volver.startsWith('/') && !volver.startsWith('//') && !volver.startsWith('/\\')) {
        return volver;
    }
    return '/';
}

// Login: el backend valida las credenciales y responde con Set-Cookie.
// El navegador guarda la cookie HttpOnly solo; nosotros no tocamos el token.
async function login(email, password) {
    const respuesta = await fetch('/signin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
        credentials: 'include',
    });

    if (respuesta.ok) {
        console.log('Login exitoso!');
        window.location.href = destino();
        return;
    }
    throw new Error(await leerError(respuesta));
}

formLogin.addEventListener('submit', async (e) => {
    e.preventDefault();
    const datos = new FormData(formLogin);
    try {
        await login(datos.get('email').trim(), datos.get('password'));
    } catch (err) {
        mostrarMensaje(err.message, 'error');
    }
});

// Registro: POST /users (es público) y después iniciamos sesión automáticamente
formRegistro.addEventListener('submit', async (e) => {
    e.preventDefault();
    const datos = new FormData(formRegistro);
    const nuevoUsuario = {
        nombre: datos.get('nombre').trim(),
        apellido: datos.get('apellido').trim(),
        dni: datos.get('dni').trim(),
        telefono: datos.get('telefono').trim(),
        email: datos.get('email').trim(),
        password: datos.get('password'),
    };

    try {
        const respuesta = await fetch('/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(nuevoUsuario),
        });
        if (!respuesta.ok) {
            // La API devuelve 500 genérico si el DNI o el email ya existen (UNIQUE)
            const detalle = await leerError(respuesta);
            throw new Error(respuesta.status === 500
                ? 'No se pudo crear la cuenta. ¿El email o el DNI ya están registrados?'
                : detalle);
        }
        await login(nuevoUsuario.email, nuevoUsuario.password);
    } catch (err) {
        mostrarMensaje(err.message, 'error');
    }
});

// Pestañas Ingresar / Crear cuenta
document.querySelectorAll('.pestania').forEach(boton => {
    boton.addEventListener('click', () => {
        const registrarse = boton.dataset.vista === 'registrarse';
        formLogin.hidden = registrarse;
        formRegistro.hidden = !registrarse;
        document.querySelectorAll('.pestania').forEach(b => b.classList.toggle('activa', b === boton));
        mostrarMensaje('', '');
    });
});

// Si ya hay una sesión activa, no tiene sentido mostrar el login
document.addEventListener('DOMContentLoaded', async () => {
    if (await checkAuthStatus()) {
        window.location.href = destino();
    }
});
