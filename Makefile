run:
	@go run .

test:
	@go test -v -count=1 ./...


chainServer:
	@cd blockchain_server && go run . -port 5000

walletServer:
	@cd wallet_server && go run . -port 8080


.PHONY: test
