test:
    docker compose down -v
	@echo "Generando sqlc y compilando"
	sqlc generate
	go build ./...
	@echo "Levantando BD"
	docker compose up -d
	@echo "Esperando a Postgres"
	@sleep 3
	@until docker compose exec db pg_isready -U postgres -d app_alquileres; do sleep 2; done
	@echo "Ejecutando tests"
	-go test -v ./... 
	docker compose down -v
