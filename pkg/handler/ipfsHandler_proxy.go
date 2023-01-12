package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"starnet/chain-api/pkg/response"
	"strings"

	"starnet/chain-api/pkg/request"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (h *IPFSHandler) bindApiKey(c echo.Context) (string, error) {
	pathList := strings.Split(c.Request().URL.Path, "/")
	if len(pathList) < 3 {
		return "", errors.New("path error")
	}
	return pathList[3], nil
}

func (h *IPFSHandler) Proxy(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.bindApiKey(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
	// 	logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
	// 	return c.JSON(http.StatusBadRequest, rlErr)
	// }

	ctx := c.Request().Context()
	w := c.Response().Writer
	r := c.Request()
	pathStr := h.ipfsMethod(apiKey, r.URL.Path)

	if !h.ipfsService.CheckMethod(pathStr) {
		// FIXME:
		// return c.JSON(http.StatusBadRequest, fmt.Errorf("not supported method"))
	}

	// Create a new HTTP client
	// TODO: remove query value
	requestURL := h.proxyURL(pathStr, r.URL.Query().Encode())

	// Create a new request to the target URL
	targetReq, err := http.NewRequest(r.Method, requestURL, r.Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
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
		return c.JSON(http.StatusBadRequest, err)
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
	w.WriteHeader(targetResp.StatusCode)

	// Copy the response body to the original response
	// TODO: read the data from the body
	body, err := io.ReadAll(targetResp.Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	_, err = w.Write(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	var bwSize int64
	var bwType uint8

	switch pathStr {
	case "/add":
		// TODO: if pin is true, call service
		addParam := new(request.AddParam)
		if err := c.Bind(addParam); err != nil {
			logger.Error("read the add param failed", zap.Error(err))
		}
		bwSize, bwType = getBwUploadParam(c, logger)
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
				}
				addResultList = append(addResultList, addResult)
			}
			logger.Info("the add results is:", zap.Any("addResultList", addResultList))
		} else {
			addResultList = append(addResultList, addResult)
			logger.Info("the add result is:", zap.Any("addResult", addResult))
		}
		//cast.ToSlice(body)
	case "/block/get":
	case "/block/put":
		bwSize, bwType = getBwUploadParam(c, logger)
	case "/block/stat":
	case "/cat":
		bwSize = int64(len(body))
		bwType = ratelimitv1.BandWidthDownload
	case "/dag/get":
	case "/dag/put":
		bwSize, bwType = getBwUploadParam(c, logger)
	case "/dag/resolve":
	case "/get":
		bwSize = int64(len(body))
		bwType = ratelimitv1.BandWidthDownload
	case "/pin/add":
		// TODO: call service
		if err := h.ipfsService.Pin(ctx, c.QueryParam("arg"), request.PinParam{}); err != nil {
			logger.Error("pin add failed", zap.Error(err))
		}
	case "/pin/ls":
		// TODO: call service
	case "/pin/rm":
	// TODO: call service

	case "/pin/update":
	case "/version":
	}

	if bwType > 0 {
		if err := h.JsonHandler.rateLimiter.BandwidthHook(c.Request().Context(), h.JsonHandler.chain.ChainID, apiKey, bwSize, bwType); err != nil {
			logger.Error("bandwidth hook failed:", zap.Error(err))
		}
	}

	return nil
}

func getBwUploadParam(c echo.Context, logger *zap.Logger) (int64, uint8) {
	requestBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("read requestBody failed", zap.Error(err))
		return 0, 0
	}
	return int64(len(requestBody)), ratelimitv1.BandWidthUpload
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
