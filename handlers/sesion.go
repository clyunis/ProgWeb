package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/lib/pq"

	db "tp5/db/sqlc"
	"tp5/views"
)

// /login : GET muestra el formulario, POST inicia sesión.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Si ya hay sesión, no tiene sentido mostrar el login
		if h.usuarioActual(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		render(w, r, http.StatusOK, views.LoginPage(views.FormLogin{}))
	case http.MethodPost:
		h.iniciarSesion(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// POST /login : valida email y contraseña, pone la cookie con el token y redirige a "/".
func (h *Handler) iniciarSesion(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}
	form := views.FormLogin{Email: strings.TrimSpace(r.FormValue("email"))}
	password := r.FormValue("password")

	if form.Email == "" || password == "" {
		form.Error = "Completá el email y la contraseña."
		render(w, r, http.StatusUnprocessableEntity, views.LoginPage(form))
		return
	}

	usuario, err := h.queries.GetUserByEmail(r.Context(), form.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Error al consultar la base de datos", http.StatusInternalServerError)
		return
	}
	// Mismo mensaje si el email no existe o si la contraseña está mal,
	// para no revelar qué emails están registrados.
	if errors.Is(err, sql.ErrNoRows) || !passwordCorrecta(usuario.Password, password) {
		form.Error = "Email o contraseña incorrectos."
		render(w, r, http.StatusUnauthorized, views.LoginPage(form))
		return
	}

	if err := emitirToken(w, usuario.IDUsuario, usuario.Email); err != nil {
		http.Error(w, "No se pudo iniciar sesión", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// /registro : GET muestra el formulario, POST crea la cuenta.
func (h *Handler) Registro(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if h.usuarioActual(r) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		render(w, r, http.StatusOK, views.RegistroPage(views.FormRegistro{}))
	case http.MethodPost:
		h.crearCuenta(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// POST /registro : crea el usuario con la contraseña hasheada, inicia sesión y redirige a "/".
func (h *Handler) crearCuenta(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}
	form := views.FormRegistro{
		Nombre:   strings.TrimSpace(r.FormValue("nombre")),
		Apellido: strings.TrimSpace(r.FormValue("apellido")),
		Dni:      strings.TrimSpace(r.FormValue("dni")),
		Telefono: strings.TrimSpace(r.FormValue("telefono")),
		Email:    strings.TrimSpace(r.FormValue("email")),
	}
	password := r.FormValue("password")

	// Los límites coinciden con los VARCHAR de la tabla USERS
	switch {
	case form.Nombre == "" || form.Apellido == "" || form.Dni == "" || form.Telefono == "" || form.Email == "":
		form.Error = "Completá todos los campos."
	case utf8.RuneCountInString(form.Nombre) > 100 || utf8.RuneCountInString(form.Apellido) > 100:
		form.Error = "El nombre y el apellido pueden tener hasta 100 caracteres."
	case len(form.Dni) > 9 || strings.Trim(form.Dni, "0123456789") != "":
		form.Error = "El DNI tiene que tener solo números (hasta 9)."
	case utf8.RuneCountInString(form.Telefono) > 20:
		form.Error = "El teléfono puede tener hasta 20 caracteres."
	case utf8.RuneCountInString(form.Email) > 100 || !strings.Contains(form.Email, "@"):
		form.Error = "Ingresá un email válido."
	case utf8.RuneCountInString(password) < 6:
		form.Error = "La contraseña tiene que tener al menos 6 caracteres."
	}
	if form.Error != "" {
		render(w, r, http.StatusUnprocessableEntity, views.RegistroPage(form))
		return
	}

	// Guardamos el hash bcrypt, nunca la contraseña en texto plano
	hash, err := hashearPassword(password)
	if err != nil {
		http.Error(w, "Error al procesar la contraseña", http.StatusInternalServerError)
		return
	}

	id, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Nombre:   form.Nombre,
		Apellido: form.Apellido,
		Dni:      form.Dni,
		Email:    form.Email,
		Telefono: form.Telefono,
		Password: hash,
	})
	if err != nil {
		// 23505 = unique_violation: ya existe un usuario con ese email o DNI
		var errPg *pq.Error
		if errors.As(err, &errPg) && errPg.Code == "23505" {
			form.Error = "Ya hay una cuenta con ese email o DNI."
			render(w, r, http.StatusConflict, views.RegistroPage(form))
			return
		}
		log.Println("Error al crear usuario:", err)
		http.Error(w, "Error al guardar en la base de datos", http.StatusInternalServerError)
		return
	}

	// Al crear la cuenta, la sesión queda iniciada
	if err := emitirToken(w, id, form.Email); err != nil {
		http.Error(w, "No se pudo iniciar sesión", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// POST /logout : borra la cookie y vuelve a la página principal.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	borrarCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
