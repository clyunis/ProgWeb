APP_NAME := tp5

.PHONY: generate build db-up db-down run test hurl clean

# Genera el código Go de las queries (sqlc) y de las vistas (templ)
generate:
	sqlc generate
	templ generate

build: generate
	@mkdir -p tmp
	go build -o tmp/$(APP_NAME) .

db-up:
	docker compose up -d
	@echo "Esperando a Postgres"
	@until docker compose exec db pg_isready -h localhost -U postgres -d app_alquileres; do sleep 2; done

db-down:
	docker compose down -v
	rm -rf uploads

run: build db-up
	./tmp/$(APP_NAME)

# Todo en un solo comando: genera, compila, levanta una base vacía, arranca el
# servidor, corre los tests y limpia (servidor, base y fotos), aunque algo falle
# o se interrumpa con Ctrl+C. Ver scripts/test.sh.
test:
	@bash scripts/test.sh

# Corre los tests contra un servidor que ya está corriendo (necesita la base vacía)
hurl:
	hurl --test requests.hurl

clean:
	rm -rf tmp
