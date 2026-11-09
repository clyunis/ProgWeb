package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Funciones de apoyo para la autenticación con JWT en una cookie HttpOnly.
// Los handlers de login, registro y logout (sesion.go) las usan.

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
	duracionToken   = 5 * time.Minute // vida de cada token
	ventanaRefresco = 4 * time.Minute // si al token le queda menos que esto, se renueva
	nombreCookie    = "token"
)

// SesionDeslizante es un middleware que envuelve a todo el servidor.
// En cada petición, si el token es válido y le queda poco tiempo de vida,
// emite uno nuevo en la misma respuesta. Así, mientras la persona use la página,
// la sesión no se cierra; si deja de usarla 5 minutos, vence.
// (En el TP4 esto lo hacía el JavaScript llamando a /refresh; sin JS, lo hace el servidor.)
func SesionDeslizante(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logout" {
			if claims, err := validarToken(r); err == nil && time.Until(claims.ExpiresAt.Time) < ventanaRefresco {
				emitirToken(w, claims.IDUsuario, claims.Email)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Claims: lo que viaja dentro del payload del JWT.
// Guardamos el id y el email del usuario (claims privados) y embebemos
// jwt.RegisteredClaims para usar los claims estándar como ExpiresAt.
type Claims struct {
	IDUsuario int32  `json:"id_usuario"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

// hashearPassword se usa al registrar un usuario: nunca guardamos la contraseña en texto plano.
func hashearPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// passwordCorrecta compara la contraseña ingresada con el hash guardado.
func passwordCorrecta(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// emitirToken crea el JWT, lo firma con jwtKey y lo envía al cliente en una cookie HttpOnly.
func emitirToken(w http.ResponseWriter, idUsuario int32, email string) error {
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
		return err
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
	return nil
}

// borrarCookie pisa la cookie con una vacía y vencida, y el navegador la elimina (logout).
func borrarCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
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
