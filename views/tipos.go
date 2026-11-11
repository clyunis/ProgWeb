package views

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

type Usuario struct {
	ID       int32
	Nombre   string
	Apellido string
}

type FormPublicacion struct {
	Titulo      string
	Descripcion string
	Calle       string
	Numero      string
	IDCiudad    int32
	Error       string
}

type FormLogin struct {
	Email string
	Error string
}

type FormRegistro struct {
	Nombre   string
	Apellido string
	Dni      string
	Telefono string
	Email    string
	Error    string
}

func idTexto(id int32) string {
	return fmt.Sprint(id)
}

func urlEliminar(id int32) string {
	return fmt.Sprintf("/publicaciones/%d/eliminar", id)
}

func textoContador(n int) string {
	if n == 1 {
		return "1 propiedad"
	}
	return fmt.Sprintf("%d propiedades", n)
}

func urlPropiedad(id int32) string {
	return fmt.Sprintf("/propiedad/%d", id)
}

func urlFoto(id int32, i, total int) string {
	i = (i%total + total) % total
	return fmt.Sprintf("/propiedad/%d?foto=%d", id, i)
}

func textoFoto(i, total int) string {
	return fmt.Sprintf("%d / %d", i+1, total)
}

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

func urlTelefono(telefono string) string {
	var digitos strings.Builder
	for _, c := range telefono {
		if unicode.IsDigit(c) || c == '+' {
			digitos.WriteRune(c)
		}
	}
	return "tel:" + digitos.String()
}

func urlConsulta(email, titulo string) string {
	asunto := url.PathEscape(fmt.Sprintf("Consulta por \"%s\" en TuAlquiler", titulo))
	return "mailto:" + email + "?subject=" + asunto
}
