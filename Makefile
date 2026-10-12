APP_NAME := tp3

.PHONY: generate build db-up db-down run hurl clean

generate:
	sqlc generate

build: generate
	@mkdir -p tmp
	go build -o tmp/$(APP_NAME) .

db-up:
	docker compose up -d
	@echo "Esperando a Postgres"
	@until docker compose exec db pg_isready -h localhost -U postgres -d app_alquileres; do sleep 2; done

db-down:
	docker compose down -v

run: build db-up
	./tmp/$(APP_NAME)

hurl:
	hurl --test requests.hurl

clean:
	rm -rf tmp
