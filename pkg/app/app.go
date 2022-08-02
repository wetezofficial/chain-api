package app

import (
	"starnet/chain-api/config"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// App 所有的依赖信息都在这里
type App struct {
	Config      *config.Config
	Logger      *zap.Logger
	Rdb         redis.UniversalClient
	HttpServer  *echo.Echo
	RateLimiter *ratelimitv1.RateLimiter

	EthHttpHandler HttpHandler
	EthWsHandler   WsHandler

	PolygonHttpHandler HttpHandler
	PolygonWsHandler   WsHandler

	ArbitrumHttpHandler HttpHandler
	ArbitrumWsHandler   WsHandler

	SolanaHttpHandler HttpHandler
	SolanaWsHandler   WsHandler

	HscHttpHandler HttpHandler
	HscWsHandler   WsHandler

	CosmosHttpHandler HttpHandler
	CosmosWsHandler   WsHandler

	EvmosHttpHandler HttpHandler
	EvmosWsHandler   WsHandler

	KavaHttpHandler HttpHandler
	KavaWsHandler   WsHandler

	// 	juno
	JunoHttpHandler HttpHandler
	JunoWsHandler   WsHandler

	// umee
	UmeeHttpHandler HttpHandler
	UmeeWsHandler   WsHandler

	GravityHttpHandler HttpHandler
	GravityWsHandler   WsHandler

	// okc
	OKCHttpHandler HttpHandler
	OKCWsHandler   WsHandler
}

func (a *App) Start() {
	err := a.HttpServer.Start(a.Config.Listen)
	if err != nil {
		a.Logger.Error("failed to run http server", zap.Error(err))
	}
}
