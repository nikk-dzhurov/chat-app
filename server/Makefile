all: build up

build:
	docker exec -t chatapp_server_1 go build -v

up:
	docker exec -t chatapp_server_1 bash -c "if pgrep app; then pkill app; fi"
	docker exec -d chatapp_server_1 ./app
