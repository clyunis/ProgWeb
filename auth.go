package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Clave secreta para firmar los tokens. Solo el servidor la conoce.
// Se lee de la variable de entorno JWT_SECRET; el valor por defecto es solo para desarrollo.
var jwtKey = []byte(claveSecreta())

func claveSecreta() string {
	if k := os.Getenv("JWT_SECRET"); k != "" {
		return k
	}
	return "my_secret_key"
}

const (
	duracionToken   = 5 * time.Minute  // vida de cada token
	ventanaRefresco = 30 * time.Second // solo se puede refrescar en los últimos 30 segundos
	nombreCookie    = "token"
)

// Claims: lo que viaja dentro del payload del JWT.
// Guardamos el id y el email del usuario (claims privados) y embebemos
// jwt.RegisteredClaims para usar los claims estándar como ExpiresAt.
type Claims struct {
	IDUsuario int32  `json:"id_usuario"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

type credenciales struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func escribirJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func escribirError(w http.ResponseWriter, status int, mensaje string) {
	escribirJSON(w, status, map[string]string{"error": mensaje})
}

// hashearPassword se usa al crear o actualizar un usuario: nunca guardamos la contraseña en texto plano.
func hashearPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// emitirToken crea el JWT, lo firma con jwtKey y lo envía al cliente en una cookie HttpOnly.
// Devuelve la fecha de expiración para informársela al frontend.
func emitirToken(w http.ResponseWriter, idUsuario int32, email string) (time.Time, error) {
	expiracion := time.Now().Add(duracionToken)
	claims := &Claims{
		IDUsuario: idUsuario,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracion),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return time.Time{}, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookie,
		Value:    tokenString,
		Path:     "/",
		Expires:  expiracion,
		HttpOnly: true, // JavaScript no puede leerla: protege el token ante XSS
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // en producción (HTTPS) la cookie solo debería viajar cifrada
	})
	return expiracion, nil
}

// validarToken lee la cookie, verifica la firma y la expiración, y devuelve los claims.
func validarToken(r *http.Request) (*Claims, error) {
	c, err := r.Cookie(nombreCookie)
	if err != nil {
		return nil, err // no hay cookie
	}

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (any, error) {
		// Rechazamos cualquier token que no esté firmado con HMAC (ej. "alg": "none")
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil || !tkn.Valid {
		return nil, errors.New("token inválido o expirado")
	}
	return claims, nil
}

// ---------- Middleware de autenticación ----------

type claveContexto struct{}

// requireAuth deja pasar la petición solo si trae un token válido.
// Los claims quedan guardados en el contexto para que el handler sepa quién es el usuario.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := validarToken(r)
		if err != nil {
			escribirError(w, http.StatusUnauthorized, "Tenés que iniciar sesión")
			return
		}
		ctx := context.WithValue(r.Context(), claveContexto{}, claims)
		next(w, r.WithContext(ctx))
	}
}

// conSesionParaEscribir: las lecturas (GET) son públicas; crear, modificar
// o borrar requiere sesión. La única excepción es POST /users (registro).
func conSesionParaEscribir(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		esLectura := r.Method == http.MethodGet || r.Method == http.MethodHead
		esRegistro := r.Method == http.MethodPost && strings.TrimSuffix(r.URL.Path, "/") == "/users"
		if esLectura || esRegistro {
			next(w, r)
			return
		}
		requireAuth(next)(w, r)
	}
}

// usuarioActual devuelve los claims del usuario logueado (solo dentro de rutas protegidas).
func usuarioActual(r *http.Request) *Claims {
	claims, _ := r.Context().Value(claveContexto{}).(*Claims)
	return claims
}

// ---------- Endpoints ----------

// POST /signin: valida email y contraseña y, si son correctos, emite el token.
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		escribirError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var creds credenciales
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		escribirError(w, http.StatusBadRequest, "Error al leer el JSON")
		return
	}
	creds.Email = strings.TrimSpace(creds.Email)
	if creds.Email == "" || creds.Password == "" {
		escribirError(w, http.StatusBadRequest, "El email y la contraseña son obligatorios")
		return
	}

	usuario, err := queries.GetUserByEmail(r.Context(), creds.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		escribirError(w, http.StatusInternalServerError, "Error al consultar la base de datos")
		return
	}
	// Mismo mensaje si el email no existe o si la contraseña está mal,
	// para no revelar qué emails están registrados.
	if errors.Is(err, sql.ErrNoRows) ||
		bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(creds.Password)) != nil {
		escribirError(w, http.StatusUnauthorized, "Email o contraseña incorrectos")
		return
	}

	expiracion, err := emitirToken(w, usuario.IDUsuario, usuario.Email)
	if err != nil {
		escribirError(w, http.StatusInternalServerError, "No se pudo generar el token")
		return
	}

	escribirJSON(w, http.StatusOK, map[string]any{
		"usuario": map[string]any{
			"id_usuario": usuario.IDUsuario,
			"nombre":     usuario.Nombre,
			"apellido":   usuario.Apellido,
			"email":      usuario.Email,
		},
		"expira": expiracion,
	})
}

// GET /welcome: verifica el token y devuelve los datos del usuario logueado.
// El frontend lo usa para saber si hay una sesión activa.
func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		escribirError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, err := validarToken(r)
	if err != nil {
		escribirError(w, http.StatusUnauthorized, "No hay una sesión válida")
		return
	}

	usuario, err := queries.GetUser(r.Context(), claims.IDUsuario)
	if err != nil {
		// El token es válido pero el usuario ya no existe (fue eliminado)
		escribirError(w, http.StatusUnauthorized, "No hay una sesión válida")
		return
	}

	escribirJSON(w, http.StatusOK, map[string]any{
		"usuario": map[string]any{
			"id_usuario": usuario.IDUsuario,
			"nombre":     usuario.Nombre,
			"apellido":   usuario.Apellido,
			"email":      usuario.Email,
		},
		"expira": claims.ExpiresAt.Time,
	})
}

// POST /refresh: si el token es válido y está por vencer, emite uno nuevo.
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		escribirError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, err := validarToken(r)
	if err != nil {
		escribirError(w, http.StatusUnauthorized, "No hay una sesión válida")
		return
	}

	// Lógica de negocio: solo se refresca cerca del vencimiento
	if time.Until(claims.ExpiresAt.Time) > ventanaRefresco {
		escribirError(w, http.StatusBadRequest, "El token todavía no está por vencer")
		return
	}

	expiracion, err := emitirToken(w, claims.IDUsuario, claims.Email)
	if err != nil {
		escribirError(w, http.StatusInternalServerError, "No se pudo generar el token")
		return
	}
	escribirJSON(w, http.StatusOK, map[string]any{"expira": expiracion})
}

// POST /logout: con JWT stateless no hay sesión que borrar en el servidor.
// Lo que hacemos es pisar la cookie con una vacía y vencida, y el navegador la elimina.
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		escribirError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}
