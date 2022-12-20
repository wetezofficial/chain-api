package appcontext

import (
	"github.com/labstack/echo/v4"
	"starnet/chain-api/pkg/app"
)

type AppContext struct {
	*app.App
	echo.Context
}
