// Referencias a los elementos del DOM que vamos a usar
const lista = document.getElementById('lista-publicaciones');
const contador = document.getElementById('contador');
const zonaPublicar = document.getElementById('zona-publicar');
const dialogo = document.getElementById('dialogo-publicar');
const form = document.getElementById('form-publicacion');
const selectCiudad = document.getElementById('select-ciudad');
const inputFotos = document.getElementById('input-fotos');
const previews = document.getElementById('previews');
const mensaje = document.getElementById('mensaje');
const btnEnviar = document.getElementById('btn-enviar');

const MAX_FOTOS = 8;
const MAX_TAMANIO = 5 * 1024 * 1024; // 5 MB, igual que el backend

// null si nadie inició sesión: la página igual muestra las publicaciones
let usuarioActual = null;

// Fotos elegidas para la publicación nueva: { archivo, url } (url = vista previa local)
let fotosElegidas = [];

// (crear, pedir y dibujarBarraSesion están en comun.js)

function mostrarMensaje(texto, tipo = '') {
    mensaje.textContent = texto;
    mensaje.className = `mensaje ${tipo}`;
}

// ---------- Cabecera según la sesión ----------

function dibujarSesion() {
    dibujarBarraSesion(usuarioActual);
    zonaPublicar.innerHTML = '';

    if (usuarioActual) {
        const btnPublicar = crear('button', 'btn-grande', '+ Publicar propiedad');
        btnPublicar.addEventListener('click', abrirDialogo);
        zonaPublicar.appendChild(btnPublicar);
    } else {
        const invitar = crear('a', 'btn-grande', '¿Tenés algo para alquilar? Ingresá y publicalo');
        invitar.href = PAGINA_LOGIN;
        zonaPublicar.appendChild(invitar);
    }
}

// ---------- Listado público ----------

// Una tarjeta por publicación. Toda la tarjeta lleva a /propiedad/{id}:
// el link está en el título y con CSS (.enlace-tarjeta::after) se estira
// para cubrir la tarjeta entera.
function crearTarjeta(pub, fotos) {
    const li = crear('li', 'tarjeta');

    const galeria = crear('div', 'galeria');
    if (fotos.length > 0) {
        const portada = crear('img', 'portada');
        portada.src = fotos[0].url;
        portada.alt = pub.titulo;
        portada.loading = 'lazy';
        galeria.appendChild(portada);
        if (fotos.length > 1) {
            galeria.appendChild(crear('span', 'cantidad-fotos', `📷 ${fotos.length}`));
        }
    } else {
        galeria.appendChild(crear('div', 'sin-foto', 'Sin fotos'));
    }

    const titulo = crear('h3');
    const enlace = crear('a', 'enlace-tarjeta', pub.titulo);
    enlace.href = `/propiedad/${pub.id_alquiler}`;
    titulo.appendChild(enlace);

    const cuerpo = crear('div', 'cuerpo');
    cuerpo.append(
        crear('p', 'ubicacion', `📍 ${pub.ciudad}, ${pub.provincia}`),
        titulo,
        crear('p', 'direccion', `${pub.direccion_calle} ${pub.numero_calle}`),
        crear('p', 'descripcion', pub.descripcion),
    );

    const pie = crear('div', 'pie');
    const esMia = usuarioActual && pub.id_usuario === usuarioActual.id_usuario;
    pie.appendChild(crear('span', 'autor', esMia ? 'Publicado por vos' : `Publicado por ${pub.autor_nombre} ${pub.autor_apellido}`));

    // Solo mostramos "Eliminar" en las publicaciones propias.
    // (El backend lo controla igual y responde 403 si no sos el dueño.)
    if (esMia) {
        const boton = crear('button', 'btn-eliminar', 'Eliminar');
        boton.type = 'button';
        boton.dataset.id = pub.id_alquiler; // el ID queda guardado en data-id
        pie.appendChild(boton);
    }
    cuerpo.appendChild(pie);

    li.append(galeria, cuerpo);
    return li;
}

// GET /publicaciones y GET /imagenes (públicos): vacía la grilla y la vuelve a dibujar
async function cargarPublicaciones() {
    try {
        const [pubs, imagenes] = await Promise.all([pedir('/publicaciones'), pedir('/imagenes')]);

        // Agrupamos las imágenes por publicación: { id_alquiler: [imagen, ...] }
        const fotosPorPub = {};
        imagenes.forEach(img => {
            (fotosPorPub[img.id_alquiler] ||= []).push(img);
        });

        lista.innerHTML = '';
        contador.textContent = pubs.length === 1 ? '1 propiedad' : `${pubs.length} propiedades`;

        if (pubs.length === 0) {
            lista.appendChild(crear('li', 'vacio', 'Todavía no hay publicaciones. ¡Sé la primera persona en publicar!'));
            return;
        }
        pubs.forEach(pub => lista.appendChild(crearTarjeta(pub, fotosPorPub[pub.id_alquiler] || [])));
    } catch (err) {
        lista.innerHTML = '';
        lista.appendChild(crear('li', 'vacio', `No se pudieron cargar las publicaciones: ${err.message}`));
    }
}

// Delegación de eventos en la grilla: un solo listener atiende
// los botones "Eliminar" de todas las tarjetas: DELETE /publicaciones/{id}
lista.addEventListener('click', async (e) => {
    const boton = e.target.closest('.btn-eliminar');
    if (!boton) return;
    if (!confirm('¿Eliminar esta publicación y sus fotos?')) return;

    boton.disabled = true;
    try {
        await pedir(`/publicaciones/${boton.dataset.id}`, { method: 'DELETE' });
        await cargarPublicaciones();
    } catch (err) {
        boton.disabled = false;
        alert(`No se pudo eliminar: ${err.message}`);
    }
});

// ---------- Formulario para publicar ----------

async function abrirDialogo() {
    mostrarMensaje('');
    dialogo.showModal();
    if (selectCiudad.options.length === 0) {
        try {
            const ciudades = await pedir('/ciudades');
            ciudades.forEach(c => {
                const opt = crear('option', '', `${c.nombre}, ${c.provincia}`);
                opt.value = c.id_ciudad;
                selectCiudad.appendChild(opt);
            });
        } catch (err) {
            mostrarMensaje(`No se pudieron cargar las ciudades: ${err.message}`, 'error');
        }
    }
}

function cerrarDialogo() {
    dialogo.close();
}

function limpiarFormulario() {
    form.reset();
    fotosElegidas.forEach(f => URL.revokeObjectURL(f.url));
    fotosElegidas = [];
    dibujarPreviews();
    mostrarMensaje('');
}

// Botones "×" y "Cancelar", y click en el fondo oscuro
dialogo.addEventListener('click', (e) => {
    if (e.target.closest('[data-cerrar]') || e.target === dialogo) cerrarDialogo();
});

// Al elegir fotos, las sumamos a las ya elegidas (se puede elegir en varias tandas)
inputFotos.addEventListener('change', () => {
    const errores = [];
    for (const archivo of inputFotos.files) {
        if (!archivo.type.startsWith('image/')) {
            errores.push(`${archivo.name} no es una imagen`);
        } else if (archivo.size > MAX_TAMANIO) {
            errores.push(`${archivo.name} pesa más de 5 MB`);
        } else if (fotosElegidas.length >= MAX_FOTOS) {
            errores.push(`Máximo ${MAX_FOTOS} fotos`);
            break;
        } else {
            fotosElegidas.push({ archivo, url: URL.createObjectURL(archivo) });
        }
    }
    inputFotos.value = ''; // permite volver a elegir el mismo archivo
    mostrarMensaje(errores.join('. '), errores.length ? 'error' : '');
    dibujarPreviews();
});

// Vistas previas locales (todavía no se subió nada al servidor)
function dibujarPreviews() {
    previews.innerHTML = '';
    fotosElegidas.forEach((foto, i) => {
        const li = crear('li', i === 0 ? 'preview principal' : 'preview');
        const img = crear('img');
        img.src = foto.url;
        img.alt = foto.archivo.name;
        const quitar = crear('button', 'quitar', '×');
        quitar.type = 'button';
        quitar.dataset.indice = i;
        quitar.setAttribute('aria-label', `Quitar ${foto.archivo.name}`);
        li.append(img, quitar);
        previews.appendChild(li);
    });
}

// Quitar una foto elegida (delegación de eventos)
previews.addEventListener('click', (e) => {
    const quitar = e.target.closest('.quitar');
    if (!quitar) return;
    const [foto] = fotosElegidas.splice(Number(quitar.dataset.indice), 1);
    URL.revokeObjectURL(foto.url);
    dibujarPreviews();
});

// Enviar: 1) POST /publicaciones (JSON)  2) POST /publicaciones/{id}/imagenes (multipart)
form.addEventListener('submit', async (e) => {
    e.preventDefault(); // evita que el navegador recargue la página

    const datos = new FormData(form);
    const nueva = {
        titulo: datos.get('titulo').trim(),
        descripcion: datos.get('descripcion').trim(),
        direccion_calle: datos.get('direccion_calle').trim(),
        numero_calle: datos.get('numero_calle').trim(),
        // Los inputs siempre devuelven strings; la API espera un número (int32).
        // No mandamos id_usuario: el backend toma al autor del token.
        id_ciudad: Number(datos.get('id_ciudad')),
    };

    btnEnviar.disabled = true;
    btnEnviar.textContent = 'Publicando…';
    try {
        const pub = await pedir('/publicaciones', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(nueva),
        });

        if (fotosElegidas.length > 0) {
            // FormData arma el multipart/form-data. No ponemos Content-Type:
            // el navegador lo completa con el "boundary" que separa los archivos.
            const fotos = new FormData();
            fotosElegidas.forEach(f => fotos.append('imagenes', f.archivo));
            try {
                await pedir(`/publicaciones/${pub.id_alquiler}/imagenes`, { method: 'POST', body: fotos });
            } catch (err) {
                // Si las fotos fallan, borramos la publicación para no dejarla a medias
                await pedir(`/publicaciones/${pub.id_alquiler}`, { method: 'DELETE' }).catch(() => {});
                throw new Error(`no se pudieron subir las fotos (${err.message})`);
            }
        }

        limpiarFormulario();
        cerrarDialogo();
        await cargarPublicaciones(); // volvemos a pedir la lista completa
    } catch (err) {
        mostrarMensaje(`No se pudo publicar: ${err.message}`, 'error');
    } finally {
        btnEnviar.disabled = false;
        btnEnviar.textContent = 'Publicar';
    }
});

// ---------- Inicio ----------

// La página es pública: si hay sesión mostramos las opciones para publicar,
// y si no, las publicaciones igual se ven.
document.addEventListener('DOMContentLoaded', async () => {
    const sesion = await checkAuthStatus();
    if (sesion) {
        usuarioActual = sesion.usuario;
        programarRefresco(sesion.expira);
    }
    dibujarSesion();
    await cargarPublicaciones();
});
