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
	docker run --restart unless-stopped --network=host --hostname rabbitmq --name rabbitmq -p 15672:15672 -p 5672:5672 -e RABBITMQ_DEFAULT_USER=user -e RABBITMQ_DEFAULT_PASS=password rabbitmq

run_redis:
	docker run --restart unless-stopped -d --network=host --memory 200m -e ALLOW_EMPTY_PASSWORD=yes redis:alpine

run_dev:
	clear && swag fmt && make update_swagger && go run cmd/api/main.go

up:
	docker-compose up --build

build:
	docker image build --network=host -t downloader_gochat .

run:
	docker run --rm --network=host --memory 300m --cpus 0.5 --name downloader_gochat --env-file ./.env downloader_gochat

push-image:
	docker tag downloader_gochat ashkanaz2828/downloader_gochat
	docker push ashkanaz2828/downloader_gochat

.PHONY: db_dev create_db_dev drop_db_dev migrate_up_db_dev migrate_down_db_dev update_swagger build_rabbitmq run_rabbitmq run_redis run_dev up build run push-image