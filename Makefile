OUTPUT=bin
FLAGS=
PROGRAMS=$(OUTPUT)/server $(OUTPUT)/client

start: fmt
	go run main/main.go
fmt:
	gofmt -s -w ./..

build: $(PROGRAMS)

$(OUTPUT)/server : server/main.go
$(OUTPUT)/client : client/main.go

$(OUTPUT)/% :
	go build -o $@ $^

clean:
	rm -r $(OUTPUT)/
