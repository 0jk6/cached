.PHONY: run build

run:
	@go run cmd/main.go $(port)

raft:
	@go run cmd/main.go $(filter-out $@,$(MAKECMDGOALS))

build:
	@go run cmd/main.go