start:
	docker compose up
build:
	docker compose up --build
start5:
	docker compose up --build --scale sensor=5 --scale mqtt_sensor=5
start25:
	docker compose up --build --scale sensor=25 --scale mqtt_sensor=5
start50:
	docker compose up --build --scale sensor=50 --scale mqtt_sensor=5
start100:
	docker compose up --build --scale sensor=100 --scale mqtt_sensor=5
fmt:
	gofmt -s -w ./..
clean:
	docker system prune
testdb:
	cd cloud_server && go test -v ./...
proto:
	sh generate_proto.sh