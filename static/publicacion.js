// Página de detalle de una publicación: /propiedad/{id}
// El servidor siempre devuelve este mismo HTML; acá leemos el id de la URL
// y pedimos los datos a la API.

const contenido = document.getElementById('contenido-detalle');
let usuarioActual = null;

function idDeLaURL() {
    const coincidencia = location.pathname.match(/^\/propiedad\/(\d+)\/?$/);
    return coincidencia ? Number(coincidencia[1]) : null;
}

function mostrarAviso(titulo, texto) {
    contenido.innerHTML = '';
    const caja = crear('div', 'aviso');
    const volver = crear('a', 'boton', 'Ver todas las publicaciones');
    volver.href = '/';
    caja.append(crear('h1', '', titulo), crear('p', '', texto), volver);
    contenido.appendChild(caja);
}

// ---------- Galería ----------

function crearGaleria(fotos, titulo) {
    const seccion = crear('section', 'galeria-detalle');

    if (fotos.length === 0) {
        seccion.appendChild(crear('div', 'sin-foto', 'Esta publicación no tiene fotos'));
        return seccion;
    }

    const visor = crear('div', 'visor');
    const enlaceGrande = crear('a'); // click en la foto: abrirla en tamaño completo
    enlaceGrande.target = '_blank';
    enlaceGrande.rel = 'noopener';
    const principal = crear('img', 'foto-principal');
    enlaceGrande.appendChild(principal);
    visor.appendChild(enlaceGrande);

    const contadorFotos = crear('span', 'contador-fotos');
    visor.appendChild(contadorFotos);

    const miniaturas = crear('div', 'tira-miniaturas');
    let actual = 0;

    function mostrar(indice) {
        actual = (indice + fotos.length) % fotos.length; // da la vuelta en los extremos
        principal.src = fotos[actual].url;
        principal.alt = `${titulo}: foto ${actual + 1} de ${fotos.length}`;
        enlaceGrande.href = fotos[actual].url;
        contadorFotos.textContent = `${actual + 1} / ${fotos.length}`;
        miniaturas.querySelectorAll('.miniatura').forEach((m, i) => {
            m.classList.toggle('activa', i === actual);
            if (i === actual) m.scrollIntoView({ block: 'nearest', inline: 'nearest' });
        });
    }

    if (fotos.length > 1) {
        const anterior = crear('button', 'flecha anterior', '‹');
        anterior.type = 'button';
        anterior.setAttribute('aria-label', 'Foto anterior');
        const siguiente = crear('button', 'flecha siguiente', '›');
        siguiente.type = 'button';
        siguiente.setAttribute('aria-label', 'Foto siguiente');
        visor.append(anterior, siguiente);

        fotos.forEach((foto, i) => {
            const boton = crear('button', 'miniatura');
            boton.type = 'button';
            boton.dataset.indice = i;
            boton.setAttribute('aria-label', `Ver foto ${i + 1}`);
            const img = crear('img');
            img.src = foto.url;
            img.alt = '';
            boton.appendChild(img);
            miniaturas.appendChild(boton);
        });

        // Delegación de eventos: un listener para flechas y miniaturas
        seccion.addEventListener('click', (e) => {
            if (e.target.closest('.anterior')) mostrar(actual - 1);
            else if (e.target.closest('.siguiente')) mostrar(actual + 1);
            else {
                const miniatura = e.target.closest('.miniatura');
                if (miniatura) mostrar(Number(miniatura.dataset.indice));
            }
        });

        // Flechas del teclado
        document.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowLeft') mostrar(actual - 1);
            if (e.key === 'ArrowRight') mostrar(actual + 1);
        });
    } else {
        contadorFotos.hidden = true;
    }

    seccion.append(visor, miniaturas);
    mostrar(0);
    return seccion;
}

// ---------- Contacto ----------

function iniciales(nombre, apellido) {
    return `${nombre[0] || ''}${apellido[0] || ''}`.toUpperCase();
}

function crearContacto(pub) {
    const aside = crear('aside', 'contacto');
    aside.appendChild(crear('h2', '', 'Contactá al dueño'));

    const persona = crear('div', 'persona');
    persona.append(
        crear('span', 'avatar', iniciales(pub.autor_nombre, pub.autor_apellido)),
        crear('span', 'nombre-dueno', `${pub.autor_nombre} ${pub.autor_apellido}`),
    );
    aside.appendChild(persona);

    const datos = crear('div', 'datos-contacto');
    aside.appendChild(datos);

    if (usuarioActual && pub.id_usuario === usuarioActual.id_usuario) {
        // Es una publicación propia
        datos.appendChild(crear('p', 'nota', 'Esta es tu publicación.'));
        const eliminar = crear('button', 'btn-eliminar', 'Eliminar publicación');
        eliminar.type = 'button';
        eliminar.addEventListener('click', async () => {
            if (!confirm('¿Eliminar esta publicación y sus fotos?')) return;
            eliminar.disabled = true;
            try {
                await pedir(`/publicaciones/${pub.id_alquiler}`, { method: 'DELETE' });
                window.location.href = '/';
            } catch (err) {
                eliminar.disabled = false;
                alert(`No se pudo eliminar: ${err.message}`);
            }
        });
        datos.appendChild(eliminar);
    } else if (!usuarioActual) {
        // Sin sesión: el email y el teléfono no se muestran
        datos.appendChild(crear('p', 'nota', 'Iniciá sesión para ver el email y el teléfono del dueño.'));
        const ingresar = crear('a', 'boton boton-ancho', 'Ingresá para ver el contacto');
        ingresar.href = urlLogin(); // después del login vuelve a esta publicación
        datos.appendChild(ingresar);
    } else {
        datos.appendChild(crear('p', 'nota', 'Cargando datos de contacto…'));
        cargarContacto(pub, datos);
    }
    return aside;
}

// GET /publicaciones/{id}/contacto (requiere sesión)
async function cargarContacto(pub, contenedor) {
    try {
        const c = await pedir(`/publicaciones/${pub.id_alquiler}/contacto`);
        contenedor.innerHTML = '';

        const email = crear('a', 'dato', `✉️ ${c.email}`);
        email.href = `mailto:${c.email}`;

        const telefono = crear('a', 'dato', `📞 ${c.telefono}`);
        telefono.href = `tel:${c.telefono.replace(/[^\d+]/g, '')}`; // tel: solo acepta dígitos y +

        const asunto = encodeURIComponent(`Consulta por "${pub.titulo}" en Techito`);
        const escribir = crear('a', 'boton boton-ancho', 'Enviar un email');
        escribir.href = `mailto:${c.email}?subject=${asunto}`;

        contenedor.append(email, telefono, escribir);
    } catch (err) {
        contenedor.innerHTML = '';
        contenedor.appendChild(crear('p', 'mensaje error', `No se pudo cargar el contacto: ${err.message}`));
    }
}

// ---------- Página completa ----------

function dibujarPublicacion(pub, fotos) {
    document.title = `${pub.titulo} · Techito`;
    contenido.innerHTML = '';

    const encabezado = crear('header', 'encabezado-detalle');
    encabezado.append(
        crear('p', 'ubicacion', `📍 ${pub.ciudad}, ${pub.provincia}`),
        crear('h1', '', pub.titulo),
        crear('p', 'direccion', `${pub.direccion_calle} ${pub.numero_calle}, ${pub.ciudad}`),
    );

    const principal = crear('div', 'columna-principal');
    const descripcion = crear('section', 'bloque');
    descripcion.append(crear('h2', '', 'Descripción'), crear('p', 'texto-largo', pub.descripcion));

    const ubicacion = crear('section', 'bloque');
    const lista = crear('dl', 'datos-ubicacion');
    [['Dirección', `${pub.direccion_calle} ${pub.numero_calle}`], ['Ciudad', pub.ciudad], ['Provincia', pub.provincia]]
        .forEach(([clave, valor]) => lista.append(crear('dt', '', clave), crear('dd', '', valor)));
    ubicacion.append(crear('h2', '', 'Ubicación'), lista);

    principal.append(crearGaleria(fotos, pub.titulo), descripcion, ubicacion);

    const layout = crear('div', 'layout-detalle');
    layout.append(principal, crearContacto(pub));

    contenido.append(encabezado, layout);
}

document.addEventListener('DOMContentLoaded', async () => {
    const id = idDeLaURL();
    if (!id) {
        mostrarAviso('Publicación no encontrada', 'La dirección no es válida.');
        return;
    }

    // Las tres cosas en paralelo: sesión (opcional), publicación e imágenes
    const [sesion, pub, fotos] = await Promise.all([
        checkAuthStatus(),
        pedir(`/publicaciones/${id}`).catch(err => err),
        pedir(`/imagenes?id_alquiler=${id}`).catch(() => []),
    ]);

    if (sesion) {
        usuarioActual = sesion.usuario;
        programarRefresco(sesion.expira);
    }
    dibujarBarraSesion(usuarioActual);

    if (pub instanceof Error) {
        if (pub.status === 404) {
            mostrarAviso('Publicación no encontrada', 'Puede que el dueño la haya eliminado.');
        } else {
            mostrarAviso('Algo salió mal', `No se pudo cargar la publicación: ${pub.message}`);
        }
        return;
    }
    dibujarPublicacion(pub, fotos);
});
