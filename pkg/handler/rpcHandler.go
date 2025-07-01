package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"starnet/chain-api/config"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/utils"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/itchyny/gojq"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type rpcNode struct {
	config.RpcNode
	Healthy atomic.Bool
}

type RpcHandler struct {
	config  *config.ChainConfig
	nodes   []*rpcNode
	jqQuery *gojq.Query
	logger  *zap.Logger
	app     *app.App
}

func NewRpcHandler(config *config.ChainConfig, logger *zap.Logger, app *app.App) (*RpcHandler, error) {
	if config.BlockNumberResultExtractor != "jq" {
		return nil, fmt.Errorf("unsupported block number result extractor: %s, only jq is supported", config.BlockNumberResultExtractor)
	}
	query, err := gojq.Parse(config.BlockNumberResultExpression)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse block number result expression")
	}

	h := &RpcHandler{
		config:  config,
		jqQuery: query,
		nodes:   make([]*rpcNode, len(config.Nodes)),
		logger:  logger,
		app:     app,
	}
	for i, node := range config.Nodes {
		h.nodes[i] = &rpcNode{
			RpcNode: node,
			Healthy: atomic.Bool{},
		}
		h.nodes[i].Healthy.Store(true)
	}

	go func() {
		for {
			h.checkNodesHealthy()
			time.Sleep(time.Minute)
		}
	}()

	return h, nil
}

func (h *RpcHandler) getHealthyNode() (*rpcNode, error) {
	for _, node := range h.nodes {
		if node.Healthy.Load() {
			return node, nil
		}
	}
	return nil, fmt.Errorf("no healthy RPC node found")
}

func (h *RpcHandler) getHealthyWsNode() (*rpcNode, error) {
	for _, node := range h.nodes {
		if node.Healthy.Load() && node.Ws != "" {
			return node, nil
		}
	}
	return nil, fmt.Errorf("no healthy ws rpc node found")
}

var internalServerError = fmt.Errorf("Internal Server Error")

func (h *RpcHandler) Http(c echo.Context) error {
	rawreq := c.Request()
	path := strings.TrimLeft(strings.TrimPrefix(rawreq.URL.Path, "/rpc/"+h.config.ChainName+"/"+h.app.RpcConfig.ApiKey), "/")
	requestID := rawreq.Context().Value("request_id").(string)
	logger := h.logger.With(zap.String("id", requestID))
	node, err := h.getHealthyNode()
	if err != nil {
		logger.Error("failed to get healthy node", zap.Error(err))
		return internalServerError
	}
	url := node.Http
	if path != "" {
		url = fmt.Sprintf("%s/%s", strings.TrimRight(node.Http, "/"), path)
	}
	fmt.Println("url", url)
	req, err := http.NewRequest(rawreq.Method, url, rawreq.Body)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return internalServerError
	}

	req.Header = rawreq.Header.Clone()
	if connectionHeader := req.Header.Get("Connection"); connectionHeader != "" {
		for _, connHeader := range strings.Split(connectionHeader, ",") {
			req.Header.Del(strings.TrimSpace(connHeader))
		}
	}
	hopByHopHeaders := []string{
		"Connection", "Keep-Alive", "Proxy-Authenticate",
		"Proxy-Authorization", "TE", "Trailer",
		"Upgrade",
	}
	for _, header := range hopByHopHeaders {
		req.Header.Del(header)
	}

	clientIP, _, err := net.SplitHostPort(rawreq.RemoteAddr)
	if err != nil {
		return err
	}
	req.Header.Set("X-Real-IP", clientIP)
	// Append to X-Forwarded-For if it already exists
	if prior, ok := rawreq.Header["X-Forwarded-For"]; ok && len(prior) > 0 {
		req.Header.Set("X-Forwarded-For", strings.Join(prior, ", ")+", "+clientIP)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("failed to do request", zap.Error(err), zap.String("url", node.Http))
		return internalServerError
	}
	defer resp.Body.Close()

	// Copy headers from response to client
	for k, v := range resp.Header {
		c.Response().Header().Set(k, strings.Join(v, ","))
	}
	c.Response().WriteHeader(resp.StatusCode)
	io.Copy(c.Response().Writer, resp.Body)

	return nil
}

func (h *RpcHandler) Ws(c echo.Context) error {
	requestID := c.Request().Context().Value("request_id").(string)
	logger := h.logger.With(zap.String("id", requestID))

	node, err := h.getHealthyWsNode()
	if err != nil {
		logger.Error("failed to get healthy ws node", zap.Error(err))
		return internalServerError
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}

	// Upgrade the HTTP connection to a WebSocket connection
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Error("failed to upgrade to websocket", zap.Error(err))
		return internalServerError
	}
	defer ws.Close()

	reqHeader := c.Request().Header.Clone()
	if connectionHeader := reqHeader.Get("Connection"); connectionHeader != "" {
		for _, connHeader := range strings.Split(connectionHeader, ",") {
			reqHeader.Del(strings.TrimSpace(connHeader))
		}
	}
	hopByHopHeaders := []string{
		"Connection", "Keep-Alive", "Proxy-Authenticate",
		"Proxy-Authorization", "TE", "Trailer",
		"Transfer-Encoding", "Upgrade",
		// WS headers
		"Sec-WebSocket-Extensions",
		"Sec-WebSocket-Key",
		"Sec-WebSocket-Version",
		"Sec-WebSocket-Protocol",
		"Sec-WebSocket-Accept",
		"Sec-WebSocket-Protocol",
	}
	for _, header := range hopByHopHeaders {
		reqHeader.Del(header)
	}

	clientIP, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		return err
	}
	reqHeader.Set("X-Real-IP", clientIP)
	// Append to X-Forwarded-For if it already exists
	if prior, ok := reqHeader["X-Forwarded-For"]; ok && len(prior) > 0 {
		reqHeader.Set("X-Forwarded-For", strings.Join(prior, ", ")+", "+clientIP)
	} else {
		reqHeader.Set("X-Forwarded-For", clientIP)
	}

	// Connect to the upstream WebSocket server
	upstream, _, err := websocket.DefaultDialer.Dial(node.Ws, reqHeader)
	if err != nil {
		logger.Error("failed to dial upstream websocket", zap.Error(err), zap.String("url", node.Ws))
		return internalServerError
	}
	defer upstream.Close()

	rawClientConn := ws.UnderlyingConn()
	rawUpstreamConn := upstream.UnderlyingConn()

	go io.Copy(rawClientConn, rawUpstreamConn)
	io.Copy(rawUpstreamConn, rawClientConn)

	return nil
}
func forward(src, dst *websocket.Conn) {
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			dst.Close()
			return
		}
		if err := dst.WriteMessage(messageType, message); err != nil {
			return
		}
	}
}

func (h *RpcHandler) checkNodesHealthy() {
	blockNumbers := make([]uint64, len(h.nodes))
	for i, node := range h.nodes {
		content := fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":[],"id":1}`, h.config.BlockNumberMethod)
		httpBlockNumber, err := getBlockNumberFromHttp(node.Http, content, h.jqQuery)
		if err != nil {
			log.Printf("failed to get block number from http %s: %v", node.Http, err)
			continue
		}
		blockNumbers[i] = httpBlockNumber

		if node.Ws == "" {
			continue
		}
		var wsBlockNumber uint64
		if h.config.ChainType == "evm" {
			wsBlockNumber, err = getBlockNumberFromEvmWs(node.Ws, content, h.jqQuery)
		} else if h.config.ChainType == "svm" {
			query, err := gojq.Parse(".params.result.slot")
			if err != nil {
				log.Printf("failed to parse block number result expression: %v", err)
				continue
			}
			wsBlockNumber, err = getBlockNumberFromSvmWs(node.Ws, query)
		}
		if err != nil {
			log.Printf("failed to get block number from ws %s: %v", node.Ws, err)
			continue
		}
		if wsBlockNumber < httpBlockNumber {
			log.Printf("ws block number %d is less than http block number %d, node %s is unhealthy", wsBlockNumber, httpBlockNumber, node.Name)
			continue
		}
		blockNumbers[i] = wsBlockNumber
	}
	maxBlockNumber := lo.Max(blockNumbers)

	for i, node := range h.nodes {
		if blockNumbers[i] == 0 || blockNumbers[i] < maxBlockNumber-h.config.MaxBehindBlocks {
			node.Healthy.Store(false)
		} else {
			node.Healthy.Store(true)
		}
	}
}

func getBlockNumberFromEvmWs(url string, content string, jqQuery *gojq.Query) (uint64, error) {
	upstream, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return 0, err
	}
	defer upstream.Close()
	if err = upstream.WriteMessage(websocket.TextMessage, []byte(content)); err != nil {
		return 0, err
	}

	for {
		messageType, message, err := upstream.ReadMessage()
		if err != nil {
			return 0, err
		}

		if messageType != websocket.TextMessage {
			continue
		}
		result, err := utils.JqQueryFirst(message, jqQuery)
		if err != nil {
			return 0, err
		}
		blockNumber, err := utils.ToUint64(result)
		if err != nil {
			return 0, err
		}
		if blockNumber == 0 {
			return 0, fmt.Errorf("main WS blockNumber is 0")
		}
		return blockNumber, nil
	}
}

func getBlockNumberFromSvmWs(url string, jqQuery *gojq.Query) (uint64, error) {
	upstream, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return 0, err
	}
	var subscriptionID int64
	defer func() {
		unsubscribeMessage := fmt.Sprintf(`{"jsonrpc": "2.0","id": 2,"method": "slotUnsubscribe", "params": [%d]}`, subscriptionID)
		// fmt.Println("unsubscribeMessage", unsubscribeMessage)
		upstream.WriteMessage(websocket.TextMessage, []byte(unsubscribeMessage))
		upstream.Close()
	}()

	if err = upstream.WriteMessage(websocket.TextMessage, []byte(`{"jsonrpc": "2.0","id": 1,"method": "slotSubscribe"}`)); err != nil {
		return 0, err
	}
	_, message, err := upstream.ReadMessage()
	if err != nil {
		return 0, err
	}
	var subscription map[string]any
	if err := json.Unmarshal(message, &subscription); err != nil {
		return 0, err
	}
	subscriptionID = int64(subscription["result"].(float64))

	for {
		messageType, message, err := upstream.ReadMessage()
		if err != nil {
			return 0, err
		}
		if messageType != websocket.TextMessage {
			continue
		}
		result, err := utils.JqQueryFirst(message, jqQuery)
		if err != nil {
			return 0, err
		}
		blockNumber, err := utils.ToUint64(result)
		if err != nil {
			return 0, err
		}
		if blockNumber == 0 {
			return 0, fmt.Errorf("main WS blockNumber is 0")
		}
		return blockNumber, nil
	}
}

func getBlockNumberFromHttp(url string, content string, jqQuery *gojq.Query) (uint64, error) {
	req, err := http.NewRequest("POST", url, io.NopCloser(bytes.NewBufferString(content)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	result, err := utils.JqQueryFirst(body, jqQuery)
	if err != nil {
		return 0, err
	}

	blockNumber, err := utils.ToUint64(result)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}
