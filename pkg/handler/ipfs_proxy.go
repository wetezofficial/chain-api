package handler

import (
	"io"
	"net/http"

	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

func (h *IPFSHandler) Proxy(c echo.Context) error {
	logger := h.newLogger(c)
	ctx := c.Request().Context()

	var err error
	errResp := map[string]interface{}{msg: ""}

	apiKey, pathStr, _, done := h.requestCheck(c, errResp, logger)
	if done {
		return nil
	}

	var bwType uint8
	var bwSize int64

	var pinAddFileList []response.AddResp
	var pinAddParam request.PinParam
	var pinUpdateParam request.UpdatePinParam

	// some user quota checks and logical preprocessing.
	switch pathStr {
	case "/add", "/dag/put", "/block/put":
		bwType = ratelimitv1.BandWidthUpload
		bwSize = cast.ToInt64(c.Request().Header[lengthHeader][0])
		_, done := h.checkIpfsUserLimit(c, apiKey, bwSize, bwType, errResp)
		if done {
			return nil
		}
	case "/dag/get", "/get", "/cat", "/block/get":
		// get object state and check user limit
		cid := c.Request().URL.Query().Get("arg")
		stats, err := h.getIPFSObjectsStats(ctx, cid)
		if err != nil {
			return h.echoReturnWithMsg(c, http.StatusInternalServerError, errResp, "get ipfs object stats error")
		}
		bwType = ratelimitv1.BandWidthDownload
		bwSize = int64(stats.CumulativeSize)
		_, done := h.checkIpfsUserLimit(c, apiKey, bwSize, bwType, errResp)
		if done {
			return nil
		}
	case "/pin/ls":
		var lsParam request.PinLsParam
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &lsParam); err != nil {
			return h.echoReturnWithMsg(c, http.StatusBadRequest, errResp, "read the pin ls param failed")
		}
		if lsParam.Arg == "" {
			return c.JSON(http.StatusOK, response.PinListMapResult{
				Keys: map[string]response.PinResp{},
			})
		}
	case "/pin/add":
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &pinAddParam); err != nil || pinAddParam.Arg == "" {
			return h.echoReturnWithMsg(c, http.StatusBadRequest, errResp, "read the pin add param failed")
		}
		// config bwType bwSize
		var shouldReturn bool
		bwType, bwSize, pinAddFileList, shouldReturn, _ = h.pinAddFile(c, apiKey, pinAddParam.Arg, logger, errResp)
		if shouldReturn {
			return nil
		}
	case "/pin/update":
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &pinUpdateParam); err != nil || len(pinUpdateParam.Arg) < 2 {
			return h.echoReturnWithMsg(c, http.StatusBadRequest, errResp, "pin update param error")
		}
		var shouldReturn bool
		bwType, bwSize, pinAddFileList, shouldReturn, _ = h.pinAddFile(c, apiKey, pinUpdateParam.Arg[1], logger, errResp)
		if shouldReturn {
			return nil
		}
	}

	// proxy the request to the ipfs
	targetResp, _, done := h.proxyToIPFS(c, pathStr, errResp)
	if done {
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("close the targetResp err:", zap.Error(err))
		}
	}(targetResp.Body)

	// Copy the response body to the original response
	body, err := io.ReadAll(targetResp.Body)
	if err != nil {
		return h.echoReturnWithMsg(c, http.StatusInternalServerError, errResp, err.Error())
	}

	// Copy the response headers to the original response
	for k, v := range targetResp.Header {
		c.Response().Writer.Header()[k] = v
	}

	// if ipfs request not success
	// fast return
	if targetResp.StatusCode != http.StatusOK {
		return h.returnIpfsResult(c, targetResp, body, errResp)
	}

	// database operation
	switch pathStr {
	case "/add":
		addResultList, err, addParam := h.getAddFileList(c, body, logger)
		go func() {
			if err = h.ipfsService.Add(ctx, apiKey, addParam, addResultList); err != nil {
				logger.Error("save upload file to database failed", zap.Error(err))
			}
		}()
	case "/pin/add":
		h.pinHook(ctx, apiKey, pinAddParam.Arg, pinAddFileList, logger)
	case "/pin/update":
		h.pinHook(ctx, apiKey, pinUpdateParam.Arg[1], pinAddFileList, logger)
		if pinUpdateParam.Unpin {
			h.unPinDBWithPined(ctx, apiKey, pinUpdateParam.Arg[0], logger)
		}
	case "/pin/rm":
		var pinRmParam request.PinRmParam
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &pinRmParam); err != nil || pinRmParam.Arg == "" {
			return h.echoReturnWithMsg(c, http.StatusBadRequest, errResp, "pin rm param error")
		}
		h.unPinDBWithPined(ctx, apiKey, pinRmParam.Arg, logger)
	}

	// update the user transfer up down usage
	if bwType > 0 {
		if err := h.JsonHandler.rateLimiter.BandwidthHook(
			ctx,
			h.JsonHandler.chain.ChainID,
			apiKey,
			bwSize,
			bwType,
		); err != nil {
			logger.Error("bandwidth hook failed:", zap.Error(err))
		}
	}

	return h.returnIpfsResult(c, targetResp, body, errResp)
}
