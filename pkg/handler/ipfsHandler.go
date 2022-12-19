package handler

import (
	"context"
	"fmt"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"starnet/chain-api/pkg/app"
	"starnet/starnet/constant"
	"time"
)

type IPFSCluster struct {
	IPFSClient  client.Client
	JsonHandler *JsonRpcHandler
}

func NewIPFSCluster(
	c client.Client,
	chain constant.Chain,
	app *app.App,
) *IPFSCluster {
	return &IPFSCluster{
		IPFSClient:  c,
		JsonHandler: NewJsonRpcHandler(chain, nil, nil, nil, nil, app),
	}
}

func (h *IPFSCluster) newLogger(c echo.Context) *zap.Logger {
	// add chain name
	return h.JsonHandler.logger.With(zap.String("chain", h.JsonHandler.chain.Name), zap.String("request_id", c.Request().Context().Value("request_id").(string)))
}

func (h *IPFSCluster) Upload(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: Upload
	fmt.Println(ctx.Err())

	// FIXME: resp cid
	return c.JSONBlob(200, nil)
}

func (h *IPFSCluster) List(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: List
	fmt.Println(ctx.Err())

	// FIXME: resp cid
	return c.JSONBlob(200, nil)
}

func (h *IPFSCluster) Get(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: Get
	fmt.Println(ctx.Err())

	return c.JSONBlob(200, nil)
}

func (h *IPFSCluster) Pin(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: Pin
	fmt.Println(ctx.Err())

	// FIXME: resp cid
	return c.JSONBlob(200, nil)
}
