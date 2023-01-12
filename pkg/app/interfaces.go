package app

import (
	"github.com/labstack/echo/v4"
)

type HttpHandler interface {
	Http(ctx echo.Context) error
	TendermintHttp(ctx echo.Context) error
}

type WsHandler interface {
	Ws(ctx echo.Context) error
}

type IPFSHandler interface {
	Add(ctx echo.Context) error
	List(ctx echo.Context) error
	Proxy(ctx echo.Context) error
}
