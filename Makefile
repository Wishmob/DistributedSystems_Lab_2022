start: fmt
	docker compose up
build: fmt
	docker compose up --build
fmt:
	gofmt -s -w ./..
clean:
	docker system prune