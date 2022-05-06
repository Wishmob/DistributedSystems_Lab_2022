OUTPUT=bin
FLAGS=
PROGRAMS=$(OUTPUT)/iot_gateway $(OUTPUT)/sensor

start: fmt
	docker compose up --build
fmt:
	gofmt -s -w ./..
