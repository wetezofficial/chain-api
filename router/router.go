package router

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"starnet/chain-api/pkg/app"
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

	return e
}
