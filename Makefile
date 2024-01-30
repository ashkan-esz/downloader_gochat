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

build_rabbitmq:
	docker image build --network=host -t rabbitmq ./docker/rabbitmq

run_rabbitmq:
	#docker run -d --hostname rabbitmq --name rabbitmq -p 15672:15672 -p 5672:5672 --network rabbitnet -e RABBITMQ_DEFAULT_USER=user -e RABBITMQ_DEFAULT_PASS=password rabbitmq
	docker run --rm --network=host --hostname rabbitmq --name rabbitmq -p 15672:15672 -p 5672:5672 -e RABBITMQ_DEFAULT_USER=user -e RABBITMQ_DEFAULT_PASS=password rabbitmq

.PHONY: db_dev create_db_dev drop_db_dev migrate_up_db_dev migrate_down_db_dev update_swagger build_rabbitmq run_rabbitmq