#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

URL="http://localhost:8080"
BINARIO="tmp/tp6"
ESPERA_BASE=60

PID_SERVIDOR=""
BASE_LEVANTADA=0

paso() { printf '\n==> %s\n' "$*"; }
error() { printf '\nERROR: %s\n' "$*" >&2; exit 1; }

limpiar() {
	local codigo=$?
	trap - EXIT INT TERM
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
	echo "Fotos borradas."

	if [[ $codigo -eq 130 || $codigo -eq 143 ]]; then
		printf '\nApp detenida.\n'
		exit 0
	fi
	if [[ $codigo -ne 0 ]]; then
		printf '\nFALLÓ (código %d).\n' "$codigo" >&2
	fi
	exit "$codigo"
}

paso "Verificando requisitos"
for herramienta in go sqlc templ docker curl; do
	command -v "$herramienta" >/dev/null || error "falta instalar '$herramienta' (ver Requisitos en el README)."
done

if curl -s -o /dev/null --max-time 2 "$URL/"; then
	error "el puerto 8080 está en uso. Detené el servidor que esté corriendo y volvé a intentar."
fi

trap limpiar EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

paso "Limpiando ejecuciones anteriores"
BASE_LEVANTADA=1
docker compose down -v >/dev/null 2>&1 || true
rm -rf uploads

paso "Generando código y compilando"
sqlc generate
templ generate
mkdir -p tmp
go build -o "$BINARIO" .

paso "Levantando la base de datos"
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

paso "Arrancando el servidor en $URL (Ctrl+C para detener y limpiar todo)"
"./$BINARIO" &
PID_SERVIDOR=$!
wait "$PID_SERVIDOR"
