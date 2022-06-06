/*
 * Created by Du, Chengbin on 2022/4/26.
 */

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io/ioutil"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/jsonrpc"
	"starnet/chain-api/pkg/proxy"
	"starnet/chain-api/pkg/utils"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"
	"starnet/starnet/constant"
	"time"
)

// JsonRpcHandler 可以处理所有使用 JsonRpc 方式通信的链，例如 ETH Polygon Arbitrum Solana
type JsonRpcHandler struct {
	chain                constant.Chain
	httpSupportedMethods []string
	wsSupportedMethods   []string
	proxy                *proxy.JsonRpcProxy
	rateLimiter          *ratelimitv1.RateLimiter
	logger               *zap.Logger
	isDev                bool
}

func NewJsonRpcHandler(
	chain constant.Chain,
	httpSupportedMethods []string,
	wsSupportedMethods []string,
	proxy *proxy.JsonRpcProxy,
	app *app.App,
) *JsonRpcHandler {
	return &JsonRpcHandler{
		chain:                chain,
		httpSupportedMethods: httpSupportedMethods,
		wsSupportedMethods:   wsSupportedMethods,
		proxy:                proxy,
		rateLimiter:          app.RateLimiter,
		logger:               app.Logger,
		isDev:                app.Config.Log.IsDevelopment,
	}
}

func (h *JsonRpcHandler) validateReq(req *jsonrpc.JsonRpcSingleRequest, supportedMethods []string) *jsonrpc.JsonRpcErr {
	if req.Method == "" {
		return jsonrpc.ParseError
	}

	if !utils.In(req.Method, supportedMethods) {
		return jsonrpc.NewUnsupportedMethodError(req.ID)
	}

	return nil
}

func (h *JsonRpcHandler) bind(rawreq []byte, supportedMethods []string) (*jsonrpc.JsonRpcRequest, *jsonrpc.JsonRpcErr) {
	req := jsonrpc.JsonRpcRequest{}
	if err := json.Unmarshal(rawreq, &req); err != nil {
		return nil, jsonrpc.ParseError
	}

	if req.IsBatchCall() {
		for _, r := range req.GetBatchCall() {
			if err := h.validateReq(&r, supportedMethods); err != nil {
				return nil, err
			}
		}
	} else {
		if err := h.validateReq(req.GetSingleCall(), supportedMethods); err != nil {
			return nil, err
		}
	}

	return &req, nil
}

func (h *JsonRpcHandler) rateLimit(ctx context.Context, logger *zap.Logger, apiKey string, n int) *jsonrpc.JsonRpcErr {
	if err := h.rateLimiter.Allow(ctx, h.chain.ChainID, apiKey, n); err != nil {
		if errors.Is(err, ratelimitv1.ExceededRateLimitError) {
			return jsonrpc.TooManyRequestErr
		}

		if errors.Is(err, ratelimitv1.ApiKeyNotExistError) {
			return jsonrpc.UnauthorizedErr
		}

		logger.Error("internal error", zap.Error(err))
		return jsonrpc.NewInternalServerError(nil)
	}
	return nil
}

func (h *JsonRpcHandler) bindApiKey(c echo.Context) (string, error) {
	var apiKey string
	err := echo.PathParamsBinder(c).String("apiKey", &apiKey).BindError()
	return apiKey, err
}

func (h *JsonRpcHandler) newLogger(c echo.Context) *zap.Logger {
	return h.logger.With(zap.String("request_id", c.Request().Context().Value("request_id").(string)))
}

func (h *JsonRpcHandler) Http(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.bindApiKey(c)
	if err != nil {
		return err
	}

	rawreq, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	logger.Debug("new request", zap.ByteString("rawreq", rawreq))
	req, vErr := h.bind(rawreq, h.httpSupportedMethods)
	if vErr != nil {
		return c.JSON(200, vErr)
	}

	if rlErr := h.rateLimit(c.Request().Context(), logger, apiKey, req.Cost()); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, _ := context.WithTimeout(c.Request().Context(), time.Second*5)
	resp, err := h.proxy.HttpProxy(ctx, logger, req)
	if err != nil {
		logger.Error("fail to proxy request", zap.Error(err))
		return c.JSON(200, jsonrpc.NewInternalServerError(nil))
	}

	logger.Debug("got response", zap.ByteString("resp", resp))

	return c.JSONBlob(200, resp)
}

func (h *JsonRpcHandler) handleWs(c echo.Context, logger *zap.Logger) error {
	upgrader := websocket.Upgrader{}
	apiKey, err := h.bindApiKey(c)
	if err != nil {
		return err
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	logger.Debug("Upgraded to WebSocket protocol")

	// 使用一个 channel 来传输数据可以解决并发写入问题
	sendCh := make(chan proxy.RespData)
	defer close(sendCh)
	go func() {
		var err error
		for {
			select {
			case resp := <-sendCh:
				if resp.Data == nil {
					return
				}

				if err = ws.WriteMessage(websocket.TextMessage, resp.Data); err != nil {
					return
				}

				if resp.Subscription {
					ctx, _ := context.WithTimeout(c.Request().Context(), time.Second*2)
					if rlErr := h.rateLimit(ctx, logger, apiKey, 1); rlErr != nil {
						logger.Warn("rate limit error", zap.Error(rlErr))
						return
					}
				}
			}
		}
	}()

	client := proxy.NewClient(ws, sendCh)
	upstreamConn, err := h.proxy.NewUpstreamWS(client)
	if err != nil {
		return err
	}
	defer upstreamConn.Close()

	resp := func(logger *zap.Logger, msg []byte) {
		logger.Debug("response", zap.ByteString("rawresp", msg))
		sendCh <- proxy.RespData{Data: msg}
	}

	respJSON := func(logger *zap.Logger, i interface{}) {
		msg, _ := json.Marshal(i)
		resp(logger, msg)
	}

	for {
		var rawreq []byte
		_, rawreq, err = ws.ReadMessage()
		if err != nil {
			logger.Debug("connection closed", zap.Error(err))
			return nil
		}

		logger.Debug("new request", zap.ByteString("rawreq", rawreq))

		req, vErr := h.bind(rawreq, h.wsSupportedMethods)
		if vErr != nil {
			respJSON(logger, vErr)
			continue
		}

		ctx, _ := context.WithTimeout(c.Request().Context(), time.Second*2)
		if rlErr := h.rateLimit(ctx, logger, apiKey, req.Cost()); rlErr != nil {
			respJSON(logger, rlErr)
			continue
		}

		ctx, _ = context.WithTimeout(c.Request().Context(), time.Second*10)
		if err = upstreamConn.Send(c.Request().Context(), logger, req); err != nil {
			logger.Error("fail to proxy request", zap.Error(err))
			respJSON(logger, jsonrpc.NewInternalServerError(nil))
			return err
		}
	}
}

func (h *JsonRpcHandler) Ws(c echo.Context) error {
	logger := h.newLogger(c).With(zap.Bool("ws", true))
	err := h.handleWs(c, logger)
	if err != nil {
		logger.Info("handle ws error", zap.Error(err))
	}
	return nil
}
