package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"starnet/starnet/constant"

	"starnet/chain-api/pkg/response"

	"starnet/chain-api/pkg/request"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	ipfsAPIKeyIndex = 3
	msg             = "message"
)

func (h *IPFSHandler) bindApiKey(c echo.Context) (string, error) {
	pathList := strings.Split(c.Request().URL.Path, "/")
	if len(pathList) < ipfsAPIKeyIndex {
		return "", errors.New("path error")
	}
	return pathList[ipfsAPIKeyIndex], nil
}

func (h *IPFSHandler) Proxy(c echo.Context) error {
	logger := h.newLogger(c)

	var err error
	errResp := map[string]interface{}{
		msg: nil,
	}

	apiKey, err := h.bindApiKey(c)
	if err != nil {
		errResp[msg] = err.Error()
		return c.JSON(http.StatusUnauthorized, errResp)
	}

	if lErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); lErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(lErr))
		errResp[msg] = "out of request limit"
		return c.JSON(http.StatusBadRequest, errResp)
	}

	ctx := c.Request().Context()
	w := c.Response().Writer
	r := c.Request()
	pathStr := h.ipfsMethod(apiKey, r.URL.Path)

	if !h.ipfsService.CheckMethod(pathStr) {
		errResp[msg] = "not supported method"
		return c.JSON(http.StatusMethodNotAllowed, errResp)
	}

	var bwType uint8

	switch pathStr {
	case "/add", "/dag/put", "/block/put":
		bwType = ratelimitv1.BandWidthUpload
	case "/dag/get", "/get", "/cat", "/block/get":
		bwType = ratelimitv1.BandWidthDownload
		if err = h.JsonHandler.rateLimiter.CheckIPFSLimit(
			ctx,
			apiKey,
			constant.ChainIPFS.ChainID,
			h.JsonHandler.logger,
			0,
			bwType,
		); err != nil {
			errResp[msg] = err.Error()
			return c.JSON(http.StatusInternalServerError, errResp)
		}
	case "/pin/ls":
		var lsParam request.PinLsParam
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &lsParam); err != nil {
			errMsg := "read the pin ls param failed"
			logger.Error(errMsg, zap.Error(err))
			errResp[msg] = errMsg
			return c.JSON(http.StatusBadRequest, errResp)
		}
		if lsParam.Arg == "" {
			return c.JSON(http.StatusOK, response.PinListMapResult{
				Keys: map[string]response.PinResp{},
			})
		}
	case "/pin/add", "/pin/rm":
		pinParam := new(request.PinParam)
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, pinParam); err != nil {
			err = fmt.Errorf("read the add param failed")
			logger.Error(err.Error(), zap.Error(err))
			errResp[msg] = err.Error()
			return c.JSON(http.StatusBadRequest, errResp)
		}
		if !h.ipfsService.CheckUserCid(ctx, apiKey, pinParam.Arg) {
			err = fmt.Errorf("can`t operation this objects")
			errResp[msg] = err.Error()
			return c.JSON(http.StatusForbidden, errResp)
		}
	case "/pin/update":
		updatePinParam := new(request.UpdatePinParam)
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, updatePinParam); err != nil {
			errMsg := "read the update param failed"
			logger.Error(errMsg, zap.Error(err))
			errResp[msg] = errMsg
			return c.JSON(http.StatusBadRequest, errResp)
		}
		for _, arg := range updatePinParam.Arg {
			if !h.ipfsService.CheckUserCid(ctx, apiKey, arg) {
				errResp[msg] = "can`t operation this objects"
				return c.JSON(http.StatusForbidden, errResp)
			}
		}
	}

	// Create a new HTTP client
	// TODO: maybe will remove same query param value
	requestURL := h.proxyURL(pathStr, r.URL.Query().Encode())

	// Create a new request to the target URL
	targetReq, err := http.NewRequest(r.Method, requestURL, r.Body)
	if err != nil {
		errResp[msg] = err.Error()
		return c.JSON(http.StatusInternalServerError, errResp)
	}

	// Copy the request headers to the new request
	targetReq.Header = r.Header
	targetReq.Header["User-Agent"] = nil
	targetReq.Header["Referer"] = nil
	targetReq.Header["Origin"] = nil

	// Forward the request to the target URL
	client := &http.Client{}
	targetResp, err := client.Do(targetReq)
	if err != nil {
		errResp[msg] = err.Error()
		return c.JSON(http.StatusBadRequest, errResp)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("close the targetResp err:", zap.Error(err))
		}
	}(targetResp.Body)

	// Copy the response headers to the original response
	for k, v := range targetResp.Header {
		w.Header()[k] = v
	}

	// Copy the response body to the original response
	body, err := io.ReadAll(targetResp.Body)
	if err != nil {
		errResp[msg] = err.Error()
		return c.JSON(http.StatusInternalServerError, errResp)
	}

	var bwSize int64

	switch pathStr {
	case "/add":
		bwSize = getBwUploadParam(c, logger)
		var addResult response.AddResp
		var addResultList []response.AddResp
		if err = json.Unmarshal(body, &addResult); err != nil {
			logger.Error("unmarshal the add result failed", zap.Error(err))
			list := strings.Split(string(body), "}")
			for _, v := range list {
				if len(v) < 5 {
					continue
				}
				if err = json.Unmarshal([]byte(v+"}"), &addResult); err != nil {
					logger.Error("unmarshal the add result failed", zap.Error(err))
					continue
				}
				addResultList = append(addResultList, addResult)
			}
		} else {
			addResultList = append(addResultList, addResult)
		}
		if err = h.ipfsService.Add(ctx, apiKey, addResultList); err != nil {
			logger.Error("save upload file to database failed", zap.Error(err))
		}
	case "/dag/get", "/get", "/cat", "/block/get":
		bwSize = int64(len(body))
	case "/dag/put", "/block/put":
		bwSize = getBwUploadParam(c, logger)
	}

	if bwType > 0 {
		if err = h.JsonHandler.rateLimiter.CheckIPFSLimit(
			ctx,
			apiKey,
			constant.ChainIPFS.ChainID,
			h.JsonHandler.logger,
			bwSize, bwType,
		); err != nil {
			errResp[msg] = err.Error()
			return c.JSON(http.StatusInternalServerError, errResp)
		}
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

	w.WriteHeader(targetResp.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		errResp[msg] = err.Error()
		return c.JSON(http.StatusInternalServerError, errResp)
	}

	return nil
}

func getBwUploadParam(c echo.Context, logger *zap.Logger) int64 {
	requestBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("read requestBody failed", zap.Error(err))
		return 0
	}
	return int64(len(requestBody))
}

func (*IPFSHandler) proxyURL(pathStr, queryStr string) string {
	requestURL := "http://127.0.0.1:5001/api/v0" + pathStr
	if queryStr != "" {
		requestURL += "?" + queryStr
	}
	return requestURL
}

func (*IPFSHandler) ipfsMethod(apiKey string, path string) string {
	pathPre := "/ipfs/v0/" + apiKey
	pathStr := strings.TrimPrefix(strings.TrimPrefix(path, pathPre), "/api/v0")
	fmt.Println(pathStr)
	return pathStr
}
