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

var jwtKey = []byte(claveSecreta())

func claveSecreta() string {
	if k := os.Getenv("JWT_SECRET"); k != "" {
		return k
	}
	return "my_secret_key"
}

const (
	duracionToken   = 5 * time.Minute
	ventanaRefresco = 4 * time.Minute
	nombreCookie    = "token"
)

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

type Claims struct {
	IDUsuario int32  `json:"id_usuario"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func hashearPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func passwordCorrecta(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

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
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

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

func validarToken(r *http.Request) (*Claims, error) {
	c, err := r.Cookie(nombreCookie)
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (any, error) {

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
