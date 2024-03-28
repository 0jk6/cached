.PHONY: run build

run:
	@go run cmd/main.go $(port)

build:
	@go run cmd/main.go