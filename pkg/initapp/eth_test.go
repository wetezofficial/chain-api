package initapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-uuid"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/big"
	"net/http"
	"net/http/httptest"
	"starnet/chain-api/config"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/initapp/chainlinktest"
	"starnet/chain-api/pkg/jsonrpc"
	"starnet/chain-api/ratelimit/v1"
	"starnet/chain-api/router"
	"starnet/starnet/dao"
	"starnet/starnet/dao/interface"
	"starnet/starnet/pkg/redis"
	"strings"
	"testing"
	"time"
)

func TestEth(t *testing.T) {
	suite.Run(t, new(ethRpcSuite))
}

type ethRpcSuite struct {
	suite.Suite
	cfg            *config.Config
	App            *app.App
	rateLimitDao   daoInterface.RateLimitDao
	httpTestServer *httptest.Server
}

func (s *ethRpcSuite) TestSingleCall() {
	apikey := s.genAndSetupApikey(10, 1000, 1, time.Now())
	c, rec := s.newHttpContext(apikey, `{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"}`)

	// Assertions
	if assert.NoError(s.T(), s.App.EthHttpHandler.Http(c)) {
		assert.Equal(s.T(), http.StatusOK, rec.Code)

		var resp jsonrpc.JsonRpcResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Nil(s.T(), err)

		assert.Greater(s.T(), len(resp.Result), 0)
		respID := (*resp.ID).(float64)
		assert.Equal(s.T(), 101, int(respID))
	}
}

func (s *ethRpcSuite) TestSingleCallRateLimit() {
	chainID := 1
	apikey := s.genAndSetupApikey(10, 1000, chainID, time.Now())
	c, _ := s.newHttpContext(apikey, `{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"}`)

	if assert.NoError(s.T(), s.App.EthHttpHandler.Http(c)) {
		usage, err := s.rateLimitDao.GetDayUsage(apikey, chainID, time.Now())
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), int64(1), usage)
	}
}

func (s *ethRpcSuite) TestBatchCallRateLimit() {
	chainID := 1
	apikey := s.genAndSetupApikey(10, 1000, chainID, time.Now())
	c, _ := s.newHttpContext(apikey, `[{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":102,"jsonrpc":"2.0"}]`)

	if assert.NoError(s.T(), s.App.EthHttpHandler.Http(c)) {
		usage, err := s.rateLimitDao.GetDayUsage(apikey, chainID, time.Now())
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), int64(2), usage)
	}
}

func (s *ethRpcSuite) TestBatchCall() {
	apikey := s.genAndSetupApikey(10, 1000, 1, time.Now())
	c, rec := s.newHttpContext(apikey, `[{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":102,"jsonrpc":"2.0"}]`)

	// Assertions
	if assert.NoError(s.T(), s.App.EthHttpHandler.Http(c)) {
		assert.Equal(s.T(), http.StatusOK, rec.Code)
		fmt.Println(rec.Body.String())
	}
}

func (s *ethRpcSuite) TestWhitelist() {
	chainID := 1
	apikey := whitelistApikey

	key := fmt.Sprintf("d:%d:{%s}:%d", chainID, apikey, time.Now().Day())
	err := s.App.Rdb.Del(context.Background(), key).Err()
	assert.Nil(s.T(), err, err)

	c, rec := s.newHttpContext(apikey, `[{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":102,"jsonrpc":"2.0"}]`)

	// Assertions
	if assert.NoError(s.T(), s.App.EthHttpHandler.Http(c)) {
		assert.Equal(s.T(), http.StatusOK, rec.Code)

		usage, err := s.rateLimitDao.GetDayUsage(apikey, chainID, time.Now())
		assert.Nil(s.T(), err, err)
		assert.Equal(s.T(), int64(2), usage)
	}
}

func (s *ethRpcSuite) TestWebSocketBatchCall() {
	apikey := s.genAndSetupApikey(10, 1000, 1, time.Now())
	ws := s.createWsConn(apikey)
	defer ws.Close()

	// write
	err := ws.WriteMessage(websocket.TextMessage, []byte(`[{"method":"eth_blockNumber","params":[],"id":1,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":1,"jsonrpc":"2.0"}]`))
	assert.Nil(s.T(), err, err)

	// read
	_, msg, err := ws.ReadMessage()
	assert.Nil(s.T(), err, err)
	var resp []jsonrpc.JsonRpcResponse
	err = json.Unmarshal(msg, &resp)
	assert.Nil(s.T(), err, err)
	fmt.Println(string(msg))
}

func (s *ethRpcSuite) TestWebSocketConcurrent() {
	secQuota := 50
	apikey := s.genAndSetupApikey(secQuota, 1000, 1, time.Now())
	ws := s.createWsConn(apikey)

	go func() {
		var err error
		for i := 0; i < secQuota; i++ {
			// write
			err = ws.WriteMessage(websocket.TextMessage, []byte(`{"method":"eth_blockNumber","params":[],"id":1,"jsonrpc":"2.0"}`))
			assert.Nil(s.T(), err, err)
		}
	}()

	for i := 0; i < secQuota; i++ {
		// read
		_, msg, err := ws.ReadMessage()
		assert.Nil(s.T(), err, err)
		var resp jsonrpc.JsonRpcResponse
		err = json.Unmarshal(msg, &resp)
		assert.Nil(s.T(), err, err)
		fmt.Println(string(msg))
	}
}

func (s *ethRpcSuite) TestWebSocket() {
	apikey := s.genAndSetupApikey(10, 1000, 1, time.Now())
	ws := s.createWsConn(apikey)

	// write
	err := ws.WriteMessage(websocket.TextMessage, []byte(`{"method":"eth_blockNumber","params":[],"id":1,"jsonrpc":"2.0"}`))
	assert.Nil(s.T(), err, err)

	// read
	_, msg, err := ws.ReadMessage()
	assert.Nil(s.T(), err, err)
	var resp jsonrpc.JsonRpcResponse
	err = json.Unmarshal(msg, &resp)
	assert.Nil(s.T(), err, err)
	fmt.Println(string(msg))
	err = ws.Close()
	assert.Nil(s.T(), err, err)
}

func (s *ethRpcSuite) TestHeadByNumber() {
	apikey := s.genAndSetupApikey(10, 1000, 1, time.Now())
	ws := s.createWsConn(apikey)

	// write
	err := ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"method":"eth_getBlockByNumber","params":["latest", false],"id":1,"jsonrpc":"2.0"}`)))
	assert.Nil(s.T(), err, err)

	// read
	_, msg, err := ws.ReadMessage()
	assert.Nil(s.T(), err, err)
	var resp jsonrpc.JsonRpcResponse
	err = json.Unmarshal(msg, &resp)
	assert.Nil(s.T(), err, err)
	fmt.Println(string(msg))

	head := new(chainlinktest.Head)
	err = json.Unmarshal(resp.Result, &head)
	assert.Nil(s.T(), err, err)

	err = ws.Close()
	assert.Nil(s.T(), err, err)
}

func (s *ethRpcSuite) TestWebsocketGetBalance() {
	apikey := s.genAndSetupApikey(10, 1000, 1, time.Now())
	ws := s.createWsConn(apikey)

	var firstTimeVal *big.Int
	for i := 0; i <= 3; i++ {
		var val *big.Int
		// write
		err := ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"method":"eth_getBalance","params":["0x5a6fCc02D8c50eA58a22115A7c4608b723030016", "latest"],"id":1,"jsonrpc":"2.0"}`)))
		assert.Nil(s.T(), err, err)

		// read
		_, msg, err := ws.ReadMessage()
		assert.Nil(s.T(), err, err)
		var resp jsonrpc.JsonRpcResponse
		err = json.Unmarshal(msg, &resp)
		assert.Nil(s.T(), err, err)
		fmt.Println(string(msg))

		var result hexutil.Big
		err = json.Unmarshal(resp.Result, &result)
		assert.Nil(s.T(), err, err)
		val = (*big.Int)(&result)
		fmt.Println(val.String())
		if i == 0 {
			firstTimeVal = val
		} else {
			assert.Equal(s.T(), firstTimeVal.String(), val.String())
		}
	}

	err := ws.Close()
	assert.Nil(s.T(), err, err)
}

const whitelistApikey = "xxx"

func (s *ethRpcSuite) SetupSuite() {
	s.loadConfig()

	logger, err := config.NewLogger(s.cfg)
	assert.Nil(s.T(), err)

	rdb := starnetRedis.New(s.cfg.Redis)

	var rateLimitDao daoInterface.RateLimitDao = dao.NewRateLimitDao(rdb)
	s.rateLimitDao = rateLimitDao

	rateLimiter, err := ratelimitv1.NewRateLimiter(rdb, logger, []string{whitelistApikey})
	assert.Nil(s.T(), err, "fail to get rate limiter")

	_app := app.App{
		Config:      s.cfg,
		Logger:      logger,
		Rdb:         rdb,
		RateLimiter: rateLimiter,
	}

	initFns := []func(app *app.App) error{
		initEthHandler,
		initPolygonHandler,
		initArbitrumHandler,
		initSolanaHandler,
	}

	for _, fn := range initFns {
		if err = fn(&_app); err != nil {
			assert.Nil(s.T(), err, "fail to run init func")
		}
	}
	_app.HttpServer = router.NewRouter(&_app)
	s.App = &_app
	s.httpTestServer = httptest.NewServer(_app.HttpServer)
}

func (s *ethRpcSuite) TearDownSuite() {
	s.httpTestServer.Close()
}

func (s *ethRpcSuite) TearDownTest() {
}

func (s *ethRpcSuite) createWsConn(apikey string) *websocket.Conn {
	wsURL := "ws" + strings.TrimPrefix(s.httpTestServer.URL, "http") + "/ws/eth/v1/" + apikey
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.Nil(s.T(), err, err)
	return ws
}

func (s *ethRpcSuite) genAndSetupApikey(secQuota, dayQuota, chainID int, t time.Time) string {
FnBegin:
	apikey, err := uuid.GenerateUUID()
	assert.Nil(s.T(), err)

	// Initialize configuration
	err = s.rateLimitDao.SetQuota(apikey, chainID, secQuota, dayQuota)
	assert.Nil(s.T(), err)

	_, err = s.rateLimitDao.GetDayUsage(apikey, chainID, t)
	if !errors.Is(err, redis.Nil) {
		// 若key已经存在，则重新生成一个作为测试
		goto FnBegin
	}

	return apikey
}

func (s *ethRpcSuite) newHttpContext(apikey string, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	requestID, _ := uuid.GenerateUUID()
	req = req.WithContext(context.WithValue(req.Context(), "request_id", requestID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := s.App.HttpServer.NewContext(req, rec)
	c.SetParamNames("apiKey")
	c.SetParamValues(apikey)

	return c, rec
}

func (s *ethRpcSuite) loadConfig() {
	cfgToml := `
listen = "127.0.0.1:1324"

[upstream]
#eth.http = "https://mainnet-rpc.wetez.io/eth/v1/a8403744cd53caeb36bc74b1978cfac2"
eth.http = "https://rinkeby-light.eth.linkpool.io"
#eth.ws = "wss://mainnet-rpc.wetez.io/ws/eth/v1/a8403744cd53caeb36bc74b1978cfac2"
eth.ws = "wss://rinkeby-light.eth.linkpool.io/ws"

arbitrum.http = "https://rinkeby.arbitrum.io/rpc"
arbitrum.ws = "wss://rinkeby-light.eth.linkpool.io/ws"

polygon.http = "https://rpc-mumbai.maticvigil.com"
polygon.ws = ""

solana.http = "https://api.testnet.solana.com"
solana.ws = ""


[[redis]]
database = 0
username = ""
password = ""
addr = "127.0.0.1:6379"

[[redis]]
database = 1
username = ""
password = ""
addr = "127.0.0.1:6379"

[log]
# debug, info, warn, error
level           = "debug"
is_dev          = true
log_file        = "stderr"

`
	viper.SetConfigType("toml")
	err := viper.ReadConfig(strings.NewReader(cfgToml))
	assert.Nil(s.T(), err)

	err = viper.Unmarshal(&s.cfg)
	assert.Nil(s.T(), err)
}
