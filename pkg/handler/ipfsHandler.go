package handler

import (
	"context"
	"encoding/json"
	"fmt"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"net/http"
	"strings"

	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"
	"starnet/starnet/models"

	ipfsApi "github.com/ipfs/go-ipfs-http-client"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"starnet/chain-api/pkg/app"
	serviceInterface "starnet/chain-api/service/interface"
	"starnet/starnet/constant"
)

type IPFSHandler struct {
	JsonHandler *JsonRpcHandler
	ipfsService serviceInterface.IpfsService
	endpoint    string
	httpClient  *http.Client
	ipfsClient  *ipfsApi.HttpApi
}

func NewIPFSHandler(
	chain constant.Chain,
	app *app.App,
) *IPFSHandler {

	ipfsHandler := &IPFSHandler{
		// without Methods & proxy
		JsonHandler: NewJsonRpcHandler(chain, nil, nil, nil, nil, nil, app),
		ipfsService: app.IPFSSrv,
		endpoint:    fmt.Sprintf("%s://%s:%d", app.Config.IPFSCluster.Schemes, app.Config.IPFSCluster.Host, app.Config.IPFSCluster.Port),
		httpClient:  &http.Client{},
	}
	node, err := ipfsApi.NewURLApiWithClient(ipfsHandler.endpoint, ipfsHandler.httpClient)
	if err != nil {
		panic(err)
	}
	ipfsHandler.ipfsClient = node
	return ipfsHandler
}

func (h *IPFSHandler) newLogger(c echo.Context) *zap.Logger {
	// add chain name
	return h.JsonHandler.logger.With(zap.String("chain", h.JsonHandler.chain.Name), zap.String("request_id", c.Request().Context().Value("request_id").(string)))
}

const (
	ipfsAPIKeyIndex = 3
	msg             = "message"
	lengthHeader    = "Content-Length"
)

func (h *IPFSHandler) bindApiKey(c echo.Context) (string, error) {
	pathList := strings.Split(c.Request().URL.Path, "/")
	if len(pathList) < ipfsAPIKeyIndex {
		return "", errors.New("path error")
	}
	if len(pathList[ipfsAPIKeyIndex-1]) != 32 {
		return "", errors.New("key error")
	}
	return pathList[ipfsAPIKeyIndex-1], nil
}

func (h *IPFSHandler) returnIpfsResult(c echo.Context, targetResp *http.Response, body []byte, errResp map[string]interface{}) error {
	w := c.Response().Writer
	w.WriteHeader(targetResp.StatusCode)
	_, err := w.Write(body)
	if err != nil {
		errResp[msg] = err.Error()
		return h.echoReturn(c, http.StatusInternalServerError, errResp)
	}

	return nil
}

func (h *IPFSHandler) checkIpfsUserLimit(c echo.Context, apiKey string, bwSize int64, bwType uint8, errResp map[string]interface{}) (error, bool) {
	ok, err := h.JsonHandler.rateLimiter.CheckIPFSLimit(
		c.Request().Context(),
		apiKey,
		constant.ChainIPFS.ChainID,
		h.JsonHandler.logger,
		bwSize,
		bwType,
	)
	if err != nil {
		errResp[msg] = err.Error()
		return h.echoReturn(c, http.StatusInternalServerError, errResp), true
	}
	if !ok {
		errResp[msg] = "out of usage limit"
		return h.echoReturn(c, http.StatusBadRequest, errResp), true
	}
	return nil, false
}

func (h *IPFSHandler) proxyURL(pathStr, queryStr string) string {
	requestURL := h.endpoint + "/api/v0" + pathStr
	if queryStr != "" {
		requestURL += "?" + queryStr
	}
	return requestURL
}

func (h *IPFSHandler) ipfsMethod(apiKey string, path string) string {
	pathStr := strings.TrimPrefix(path, "/ipfs")
	pathStr = strings.TrimPrefix(pathStr, "/"+apiKey)
	pathStr = strings.TrimPrefix(pathStr, "/api/v0")
	return pathStr
}

func (h *IPFSHandler) getAddFileList(c echo.Context, body []byte, logger *zap.Logger) ([]response.AddResp, error, request.AddParam) {
	var addResult response.AddResp
	var addResultList []response.AddResp
	var err error
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
	var addParam request.AddParam
	if err := (&echo.DefaultBinder{}).BindQueryParams(c, &addParam); err != nil {
		logger.Error("read the add param failed", zap.Error(err))
	}
	// add param
	addParam.PinStatus = models.PinStatusPin
	pinedStr := c.Request().URL.Query().Get("pin")
	if pinedStr == "false" {
		addParam.PinStatus = models.PinStatusUnPin
	}
	return addResultList, err, addParam
}

func (h *IPFSHandler) proxyToIPFS(c echo.Context, pathStr string, errResp map[string]interface{}) (*http.Response, error, bool) {
	r := c.Request()
	requestURL := h.proxyURL(pathStr, r.URL.Query().Encode())

	// Create a new request to the target URL
	targetReq, err := http.NewRequest(r.Method, requestURL, r.Body)
	if err != nil {
		errResp[msg] = err.Error()
		return nil, h.echoReturn(c, http.StatusInternalServerError, errResp), true
	}

	// Copy the request headers to the new request
	targetReq.Header = r.Header
	targetReq.Header["User-Agent"] = nil
	targetReq.Header["Referer"] = nil
	targetReq.Header["Origin"] = nil

	// Forward the request to the target URL
	targetResp, err := h.httpClient.Do(targetReq)
	if err != nil {
		errResp[msg] = err.Error()
		return nil, h.echoReturn(c, http.StatusBadRequest, errResp), true
	}

	return targetResp, nil, false
}

func (h *IPFSHandler) requestCheck(c echo.Context, errResp map[string]interface{}, logger *zap.Logger) (string, string, error, bool) {
	var apiKey string
	var err error
	r := c.Request()
	subdomain := strings.Split(r.Header.Get("Domain"), ".")[0]
	if subdomain == "" {
		apiKey, err = h.bindApiKey(c)
		if err != nil {
			errResp[msg] = err.Error()
			return "", "", h.echoReturn(c, http.StatusUnauthorized, errResp), true
		}
	} else {
		apiKey = h.ipfsService.GetApiKeyByActiveGateway(r.Context(), subdomain)
		if apiKey == "" {
			errResp[msg] = "not found apiKey"
			return "", "", h.echoReturn(c, http.StatusUnauthorized, errResp), true
		}
	}

	pathStr := h.ipfsMethod(apiKey, r.URL.Path)

	if strings.Contains(pathStr, "ping") {
		return "", "", c.String(http.StatusOK, "pong"), true
	}

	if !h.ipfsService.CheckMethod(pathStr) {
		errResp[msg] = "not supported method"
		return "", "", h.echoReturn(c, http.StatusMethodNotAllowed, errResp), true
	}

	if rlErr := h.JsonHandler.rateLimit(r.Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		errResp[msg] = rlErr.Error()
		return "", "", h.echoReturn(c, http.StatusBadRequest, errResp), true
	}

	return apiKey, pathStr, nil, false
}

func (h *IPFSHandler) pinAddObject(ctx context.Context, cid string) error {
	p := path.New(cid)
	return h.ipfsClient.Pin().Add(ctx, p)
}

func (h *IPFSHandler) getIPFSObjectsStats(ctx context.Context, cid string) (*iface.ObjectStat, error) {
	p := path.New(cid)
	stat, err := h.ipfsClient.Object().Stat(ctx, p)
	if err != nil {
		return nil, err
	}
	return stat, nil
}

func (h *IPFSHandler) pinAddFile(c echo.Context, apiKey, cid string, logger *zap.Logger, errResp map[string]interface{}) (uint8, int64, []response.AddResp, bool, error) {
	var pinAddFileList []response.AddResp
	var bwType uint8
	var bwSize int64
	file, _ := h.ipfsService.GetUserIpfsFile(c.Request().Context(), apiKey, cid)
	if file.ID == 0 {
		stats, err := h.getIPFSObjectsStats(c.Request().Context(), cid)
		if err != nil {
			errMsg := "pin files failed"
			logger.Error(errMsg, zap.Error(err))
			errResp[msg] = errMsg
			return 0, 0, nil, true, h.echoReturn(c, http.StatusInternalServerError, errResp)
		}

		bwType = ratelimitv1.BandWidthUpload
		bwSize := int64(stats.CumulativeSize)

		_, done := h.checkIpfsUserLimit(c, apiKey, bwSize, bwType, errResp)
		if done {
			return 0, 0, nil, true, nil
		}

		pinAddFileList = append(pinAddFileList, response.AddResp{
			Hash: stats.Cid.String(),
			Size: cast.ToString(stats.CumulativeSize),
		})
	}
	return bwType, bwSize, pinAddFileList, false, nil
}

func (h *IPFSHandler) pinHook(ctx context.Context, apiKey, pinCid string, pinAddFileList []response.AddResp, logger *zap.Logger) {
	if len(pinAddFileList) > 0 {
		if err := h.ipfsService.Add(ctx, apiKey, request.AddParam{
			PinStatus: models.PinStatusPin,
		}, pinAddFileList); err != nil {
			logger.Error("save pin upload file to database failed", zap.Error(err))
		}
	} else {
		if err := h.ipfsService.PinObject(ctx, apiKey, pinCid); err != nil {
			logger.Error("save pin status file to database failed", zap.Error(err))
		}
	}
}

func (h *IPFSHandler) echoReturnWithMsg(c echo.Context, httpStatus int, resp map[string]interface{}, msg string) error {
	resp[msg] = msg
	return h.echoReturn(c, httpStatus, resp)
}

func (h *IPFSHandler) echoReturn(c echo.Context, httpStatus int, resp map[string]interface{}) error {
	logger := h.newLogger(c)
	if httpStatus != http.StatusOK {
		logger.Error("request failed", zap.Int("httpStatus", httpStatus), zap.Any("resp", resp))
	}
	return c.JSON(httpStatus, resp)
}

func (h *IPFSHandler) unPinDBWithPined(ctx context.Context, apiKey string, cid string, logger *zap.Logger) {
	_ = h.ipfsService.UnPinObject(ctx, apiKey, cid)
	file, _ := h.ipfsService.GetIpfsPinFile(ctx, cid)
	if file.ID > 0 {
		if err := h.pinAddObject(ctx, cid); err != nil {
			logger.Error("pin object failed", zap.Error(err))
		}
	}
}
