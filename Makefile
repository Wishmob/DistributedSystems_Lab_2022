start: fmt
	docker compose up

build: fmt
	docker compose up --build
start5:
	docker compose up --build --scale sensor=5
start25:
	docker compose up --build --scale sensor=25
start50:
	docker compose up --build --scale sensor=50
start100:
	docker compose up --build --scale sensor=100
fmt:
	gofmt -s -w ./..
clean:
	docker system prune