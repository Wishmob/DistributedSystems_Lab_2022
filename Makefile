start: fmt
	docker compose up --build
fmt:
	gofmt -s -w ./..
