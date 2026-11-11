#!/usr/bin/env bash
#
# Corre los tests de punta a punta con un solo comando (make test):
#
#   1. verifica que estén las herramientas y que el puerto 8080 esté libre
#   2. genera el código (sqlc + templ) y compila
#   3. levanta una base de datos vacía con Docker y espera a que esté lista
#   4. arranca el servidor y espera a que responda
#   5. corre las pruebas de requests.hurl
#   6. limpia SIEMPRE: apaga el servidor, baja la base y borra las fotos subidas,
#      aunque algo falle o se interrumpa con Ctrl+C
#
# Termina con código 0 si todo pasó, y distinto de 0 si algo falló.

set -euo pipefail

# Trabajamos siempre desde la raíz del proyecto
cd "$(dirname "$0")/.."

URL="http://localhost:8080"
BINARIO="tmp/tp5"
LOG="tmp/server.log"
ESPERA_BASE=60     # segundos máximos para que Postgres esté listo
ESPERA_SERVIDOR=20 # segundos máximos para que el servidor responda

PID_SERVIDOR=""
BASE_LEVANTADA=0

paso() { printf '\n==> %s\n' "$*"; }
error() { printf '\nERROR: %s\n' "$*" >&2; exit 1; }

# ---------- Limpieza: se ejecuta siempre al salir ----------

limpiar() {
	local codigo=$?
	trap - EXIT
	paso "Limpiando"

	if [[ -n "$PID_SERVIDOR" ]] && kill -0 "$PID_SERVIDOR" 2>/dev/null; then
		kill "$PID_SERVIDOR" 2>/dev/null || true
		wait "$PID_SERVIDOR" 2>/dev/null || true
		echo "Servidor detenido."
	fi

	if [[ "$BASE_LEVANTADA" == 1 ]]; then
		docker compose down -v >/dev/null 2>&1 || true
		echo "Base de datos bajada."
	fi

	rm -rf uploads
	echo "Fotos de prueba borradas."

	if [[ $codigo -eq 0 ]]; then
		printf '\nOK: todos los tests pasaron.\n'
	else
		if [[ -s "$LOG" ]]; then
			printf '\n--- Últimas líneas del log del servidor (%s) ---\n' "$LOG"
			tail -n 20 "$LOG"
		fi
		if [[ $codigo -eq 130 ]]; then
			printf '\nInterrumpido con Ctrl+C.\n' >&2
		else
			printf '\nFALLÓ (código %d).\n' "$codigo" >&2
		fi
	fi
	exit "$codigo"
}

# ---------- 1. Verificaciones previas ----------

paso "Verificando requisitos"
for herramienta in go sqlc templ docker hurl curl; do
	command -v "$herramienta" >/dev/null || error "falta instalar '$herramienta' (ver Requisitos en el README)."
done

# Si el 8080 ya responde (por ejemplo, quedó corriendo 'make run'), no tocamos nada:
# los tests correrían contra ese servidor y no contra el que compilamos.
if curl -s -o /dev/null --max-time 2 "$URL/"; then
	error "el puerto 8080 está en uso. Detené el servidor que esté corriendo y volvé a intentar."
fi

# A partir de acá, cualquier salida (error, tests que fallan o Ctrl+C) pasa por limpiar.
trap limpiar EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# ---------- 2. Generar y compilar ----------

paso "Generando código y compilando"
sqlc generate
templ generate
mkdir -p tmp
go build -o "$BINARIO" .

# ---------- 3. Base de datos vacía ----------

paso "Levantando una base de datos vacía"
BASE_LEVANTADA=1
docker compose down -v >/dev/null 2>&1 || true
rm -rf uploads
docker compose up -d

echo "Esperando a Postgres (máximo ${ESPERA_BASE}s)..."
for ((i = 0; i < ESPERA_BASE; i++)); do
	if docker compose exec -T db pg_isready -h localhost -U postgres -d app_alquileres >/dev/null 2>&1; then
		echo "Postgres listo."
		break
	fi
	sleep 1
done
((i < ESPERA_BASE)) || error "Postgres no estuvo listo en ${ESPERA_BASE}s."

# ---------- 4. Servidor ----------

paso "Arrancando el servidor (log en $LOG)"
"./$BINARIO" >"$LOG" 2>&1 &
PID_SERVIDOR=$!

for ((i = 0; i < ESPERA_SERVIDOR * 2; i++)); do
	kill -0 "$PID_SERVIDOR" 2>/dev/null || error "el servidor se detuvo al arrancar."
	if curl -s -o /dev/null --max-time 1 "$URL/"; then
		echo "Servidor listo."
		break
	fi
	sleep 0.5
done
((i < ESPERA_SERVIDOR * 2)) || error "el servidor no respondió en ${ESPERA_SERVIDOR}s."

# ---------- 5. Tests ----------

paso "Corriendo los tests (requests.hurl)"
hurl --test requests.hurl
