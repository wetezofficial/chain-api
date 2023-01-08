module starnet/chain-api

go 1.16

replace starnet/starnet => ../starnet

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.18
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/go-uuid v1.0.1
	github.com/jonboulle/clockwork v0.3.0 // indirect
	github.com/labstack/echo/v4 v4.7.2
	github.com/pkg/errors v0.9.1
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.2
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/zap v1.21.0
	starnet/starnet v0.0.0-00010101000000-000000000000
)
