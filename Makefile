
service-db-up:
	@docker compose up -d db

service-up:
	@docker compose up

service-down:
	@docker compose down

service-up-rebuild:
	@docker compose up --build