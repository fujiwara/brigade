CLIENTS := 8
OPTIONS := ""

.PHONY: clean up down all docker-build

all: brigade docker-compose.yml docker-build

docker-build: Dockerfile
	docker build -t brigade:test .

clean: down
	rm -f brigade docker-compose.yml

brigade: ../brigade.go ../cmd/brigade/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ../cmd/brigade

docker-compose.yml: docker-compose.yml.erb
	CLIENTS=$(CLIENTS) erb docker-compose.yml.erb > docker-compose.yml

up: all
	docker-compose up

down:
	docker-compose down
