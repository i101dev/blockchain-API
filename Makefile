run:
	@go run .

test:
	@go test -v -count=1 ./...


chainServer:
	@cd blockchain_server && go run . -port 5000
chainServer1:
	@cd blockchain_server && go run . -port 5001
chainServer2:
	@cd blockchain_server && go run . -port 5002

walletServer:
	@cd wallet_server && go run . -port 8080
walletServer1:
	@cd wallet_server && go run . -port 8081


.PHONY: test
