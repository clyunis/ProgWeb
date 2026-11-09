package views

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// Usuario es el usuario con sesión iniciada que necesitan las vistas.
// Si nadie inició sesión, las vistas reciben nil.
type Usuario struct {
	ID       int32
	Nombre   string
	Apellido string
}

// FormPublicacion son los valores del formulario de publicar. Cuando la
// validación falla, se vuelve a mostrar el formulario con lo que la persona
// ya había escrito y el mensaje de error.
type FormPublicacion struct {
	Titulo      string
	Descripcion string
	Calle       string
	Numero      string
	IDCiudad    int32
	Error       string
}

// FormLogin son los valores del formulario de ingreso.
// La contraseña nunca se vuelve a mostrar.
type FormLogin struct {
	Email string
	Error string
}

// FormRegistro son los valores del formulario para crear una cuenta.
type FormRegistro struct {
	Nombre   string
	Apellido string
	Dni      string
	Telefono string
	Email    string
	Error    string
}

// idTexto convierte un id numérico a texto (para el value de un <option>).
func idTexto(id int32) string {
	return fmt.Sprint(id)
}

// urlEliminar arma la ruta del formulario que elimina una publicación.
func urlEliminar(id int32) string {
	return fmt.Sprintf("/publicaciones/%d/eliminar", id)
}

// textoContador devuelve "1 propiedad" o "N propiedades".
func textoContador(n int) string {
	if n == 1 {
		return "1 propiedad"
	}
	return fmt.Sprintf("%d propiedades", n)
}

// urlPropiedad arma el link a la página de detalle de una publicación.
func urlPropiedad(id int32) string {
	return fmt.Sprintf("/propiedad/%d", id)
}

// urlFoto arma el link a la página de detalle mostrando la foto número i.
// Si i se pasa del rango, da la vuelta (después de la última viene la primera).
func urlFoto(id int32, i, total int) string {
	i = (i%total + total) % total
	return fmt.Sprintf("/propiedad/%d?foto=%d", id, i)
}

// textoFoto devuelve "2 / 5".
func textoFoto(i, total int) string {
	return fmt.Sprintf("%d / %d", i+1, total)
}

// iniciales devuelve "AP" para "Ana Pérez".
func iniciales(nombre, apellido string) string {
	var r string
	for _, s := range []string{nombre, apellido} {
		for _, c := range s {
			r += strings.ToUpper(string(c))
			break
		}
	}
	return r
}

// urlTelefono arma un link tel: (solo acepta dígitos y +).
func urlTelefono(telefono string) string {
	var digitos strings.Builder
	for _, c := range telefono {
		if unicode.IsDigit(c) || c == '+' {
			digitos.WriteRune(c)
		}
	}
	return "tel:" + digitos.String()
}

// urlConsulta arma un link mailto: con el asunto ya escrito.
func urlConsulta(email, titulo string) string {
	asunto := url.PathEscape(fmt.Sprintf("Consulta por \"%s\" en TuAlquiler", titulo))
	return "mailto:" + email + "?subject=" + asunto
}
