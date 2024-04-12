all: compile

compile:
	echo "Compiling for linux"
	cd cmd/banner && GOOS=linux GOARCH=amd64 go build -o ../../banner.elf .

run:
	./banner.elf

test:
	go test ./...

docker-up:
	docker-compose -f docker-compose.dev.yaml up -d

docker-down:
	docker-compose -f docker-compose.dev.yaml down

docker-build:
	docker-compose -f docker-compose.dev.yaml up --build

docker-restart:
	docker-compose -f docker-compose.dev.yaml restart

up: compile run

restart: docker-restart run

