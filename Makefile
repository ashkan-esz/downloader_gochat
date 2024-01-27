db_dev:
	docker run --restart unless-stopped --network=host --memory 500m -v pgdata:/var/lib/postgresql/data -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=downloader postgres:16.0-alpine3.18

create_db_dev:
	docker exec -it postgres16 createdb --username=root --owner=root go=chat

drop_db_dev:
	docker rexec -it postgres16 dropdb go-chat

migrate_up_db_dev:
	migrate -path db/migrations -database "postgres://root:mysecretpassword@localhost:5432/go-chat?sslmode=disable" -verbose up

migrate_down_db_dev:
	migrate -path db/migrations -database "postgres://root:mysecretpassword@localhost:5432/go-chat?sslmode=disable" -verbose down

update_swagger:
	swag init -g cmd/api/main.go --output docs

.PHONY: db_dev create_db_dev drop_db_dev migrate_up_db_dev migrate_down_db_dev update_swagger