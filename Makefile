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

# Corre los tests sobre una base limpia: borra la base, la levanta vacía,
# arranca el servidor, ejecuta requests.hurl y apaga el servidor.
test: build
	docker compose down -v
	rm -rf uploads
	$(MAKE) db-up
	@./tmp/$(APP_NAME) & echo $$! > tmp/server.pid; \
	until curl -s -o /dev/null http://localhost:8080/; do sleep 0.5; done; \
	hurl --test requests.hurl; resultado=$$?; \
	kill $$(cat tmp/server.pid); rm -f tmp/server.pid; \
	exit $$resultado

# Corre los tests contra un servidor que ya está corriendo (necesita la base vacía)
hurl:
	hurl --test requests.hurl

clean:
	rm -rf tmp
