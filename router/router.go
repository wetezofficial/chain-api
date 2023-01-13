package router

import (
	"context"

	"starnet/chain-api/pkg/app"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(app *app.App) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c echo.Context, requestID string) {
			ctx := context.WithValue(c.Request().Context(), "request_id", requestID)
			c.SetRequest(c.Request().WithContext(ctx))
		},
	}))

	e.POST("/eth/v1/:apiKey", app.EthHttpHandler.Http)
	e.GET("/ws/eth/v1/:apiKey", app.EthWsHandler.Ws)

	e.POST("/polygon/v1/:apiKey", app.PolygonHttpHandler.Http)
	e.GET("/ws/polygon/v1/:apiKey", app.PolygonWsHandler.Ws)

	e.POST("/arbitrum/v1/:apiKey", app.ArbitrumHttpHandler.Http)
	e.GET("/ws/arbitrum/v1/:apiKey", app.ArbitrumWsHandler.Ws)

	e.POST("/solana/v1/:apiKey", app.SolanaHttpHandler.Http)
	e.GET("/ws/solana/v1/:apiKey", app.SolanaWsHandler.Ws)

	e.POST("/hsc/v1/:apiKey", app.HscHttpHandler.Http)
	e.GET("/ws/hsc/v1/:apiKey", app.HscWsHandler.Ws)

	e.POST("/cosmos/tendermint/v1/:apiKey", app.CosmosHttpHandler.Http)
	e.GET("/cosmos/tendermint/v1/:apiKey", app.CosmosHttpHandler.TendermintHttp)
	e.GET("/ws/cosmos/tendermint/v1/:apiKey", app.CosmosWsHandler.Ws)

	e.POST("/evmos/tendermint/v1/:apiKey", app.EvmosHttpHandler.Http)
	e.GET("/evmos/tendermint/v1/:apiKey", app.EvmosHttpHandler.TendermintHttp)
	e.GET("/ws/evmos/tendermint/v1/:apiKey", app.EvmosWsHandler.Ws)

	e.POST("/kava/tendermint/v1/:apiKey", app.KavaHttpHandler.Http)
	e.GET("/kava/tendermint/v1/:apiKey", app.KavaHttpHandler.TendermintHttp)
	e.GET("/ws/kava/tendermint/v1/:apiKey", app.KavaWsHandler.Ws)

	e.POST("/juno/tendermint/v1/:apiKey", app.JunoHttpHandler.Http)
	e.GET("/juno/tendermint/v1/:apiKey", app.JunoHttpHandler.TendermintHttp)
	e.GET("/ws/juno/tendermint/v1/:apiKey", app.JunoWsHandler.Ws)

	e.POST("/umee/tendermint/v1/:apiKey", app.UmeeHttpHandler.Http)
	e.GET("/umee/tendermint/v1/:apiKey", app.UmeeHttpHandler.TendermintHttp)
	e.GET("/ws/umee/tendermint/v1/:apiKey", app.UmeeWsHandler.Ws)

	e.POST("/gravity/tendermint/v1/:apiKey", app.GravityHttpHandler.Http)
	e.GET("/gravity/tendermint/v1/:apiKey", app.GravityHttpHandler.TendermintHttp)
	e.GET("/ws/gravity/tendermint/v1/:apiKey", app.GravityWsHandler.Ws)

	e.POST("/okc/tendermint/v1/:apiKey", app.OKCHttpHandler.Http)
	e.GET("/okc/tendermint/v1/:apiKey", app.OKCHttpHandler.TendermintHttp)
	e.GET("/ws/okc/tendermint/v1/:apiKey", app.OKCWsHandler.Ws)

	e.POST("/irisnet/tendermint/v1/:apiKey", app.IRISnetHttpHandler.Http)
	e.GET("/irisnet/tendermint/v1/:apiKey", app.IRISnetHttpHandler.TendermintHttp)
	e.GET("/ws/irisnet/tendermint/v1/:apiKey", app.IRISnetWsHandler.Ws)

	e.Any("/ipfs/v0/:apiKey/*", app.IPFSHandler.Proxy)

	return e
}
