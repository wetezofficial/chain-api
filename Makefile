.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/prod/starnet-chain-api -ldflags="-s -w" -trimpath ./cmd/chainapi

.PHONY: buildTest
buildTest:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/test/starnet-chain-api -gcflags "all=-N -l" ./cmd/chainapi

.PHONY: deployTest
deployTest:
	scp ./build/test/starnet-chain-api starnettestserver:
	ssh starnettestserver "systemctl stop starnet-chain-api && cp /root/starnet-chain-api /usr/local/bin/starnet-chain-api && systemctl restart starnet-chain-api"
	@echo "Success to deploy on test server"
