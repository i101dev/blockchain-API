run:
	@go run .

test:
	@go test -v -count=1 ./...

chainServer:
	@go run blockchain_server/main.go

.PHONY: test
