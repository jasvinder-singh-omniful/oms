APP_NAME := oms
BIN_PATH := ./tmp/main
SRC_PATH := ./cmd/server
CONFIG := config/local.yml


# test makefile formatting of tabs (^I denotes tab and $ denotes line ending)
check:
	cat -e -t -v Makefile

# build go binary 
build:
	go build -o $(BIN_PATH) $(SRC_PATH)

# run the dev server
run-dev:
	docker stop postgres || true
	docker stop redis || true
	docker compose -f docker/docker-compose.yml up -d
	air

run-test:
	docker stop postgres || true
	docker stop redis || true
	docker compose -f docker/docker-compose.yml up -d
	go run $(SRC_PATH)

# clean up binary
clean:
	rm -f $(BIN_PATH)

# run tests
test:
	go test ./...


.PHONY: build run clean test check

