/*
 * Created by Chengbin Du on 2022/6/9.
 */

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
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"starnet/chain-api/config"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/initapp/chainlinktest"
	"starnet/chain-api/pkg/jsonrpc"
	"starnet/chain-api/ratelimit/v1"
	"starnet/chain-api/router"
	"starnet/starnet/constant"
	"starnet/starnet/dao"
	"starnet/starnet/dao/interface"
	"starnet/starnet/pkg/redis"
	"strings"
	"testing"
	"time"
)

func TestHsc(t *testing.T) {
	suite.Run(t, new(hscRpcSuite))
}

type hscRpcSuite struct {
	suite.Suite
	cfg            *config.Config
	App            *app.App
	rateLimitDao   daoInterface.RateLimitDao
	httpTestServer *httptest.Server
	chainID        int
}

func (s *hscRpcSuite) TestSingleCall() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
	respBytes := s.httpRequest(apikey, `{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"}`)

	var resp jsonrpc.JsonRpcResponse
	err := json.Unmarshal(respBytes, &resp)
	assert.Nil(s.T(), err)

	assert.Greater(s.T(), len(resp.Result), 0)
	respID := (*resp.ID).(float64)
	assert.Equal(s.T(), 101, int(respID))
}

func (s *hscRpcSuite) TestSingleCallRateLimit() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
	_ = s.httpRequest(apikey, `{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"}`)

	usage, err := s.rateLimitDao.GetDayUsage(apikey, int(s.chainID), time.Now())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(1), usage)
}

func (s *hscRpcSuite) TestBatchCallRateLimit() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
	_ = s.httpRequest(apikey, `[{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":102,"jsonrpc":"2.0"}]`)

	usage, err := s.rateLimitDao.GetDayUsage(apikey, int(s.chainID), time.Now())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int64(2), usage)
}

func (s *hscRpcSuite) TestBatchCall() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
	respBytes := s.httpRequest(apikey, `[{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":102,"jsonrpc":"2.0"}]`)

	fmt.Println(string(respBytes))
}

func (s *hscRpcSuite) TestWhitelist() {
	apikey := whitelistApikey

	key := fmt.Sprintf("d:%d:{%s}:%d", s.chainID, apikey, time.Now().Day())
	err := s.App.Rdb.Del(context.Background(), key).Err()
	assert.Nil(s.T(), err, err)

	_ = s.httpRequest(apikey, `[{"method":"eth_blockNumber","params":[],"id":101,"jsonrpc":"2.0"},{"method":"eth_blockNumber","params":[],"id":102,"jsonrpc":"2.0"}]`)

	usage, err := s.rateLimitDao.GetDayUsage(apikey, s.chainID, time.Now())
	assert.Nil(s.T(), err, err)
	assert.Equal(s.T(), int64(2), usage)
}

func (s *hscRpcSuite) TestWebSocketBatchCall() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
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

func (s *hscRpcSuite) TestWebSocketConcurrent() {
	secQuota := 50
	apikey := s.genAndSetupApikey(secQuota, 1000, time.Now())
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

func (s *hscRpcSuite) TestWebSocket() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
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

func (s *hscRpcSuite) TestHeadByNumber() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
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

func (s *hscRpcSuite) TestWebsocketGetBalance() {
	apikey := s.genAndSetupApikey(10, 1000, time.Now())
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

func (s *hscRpcSuite) SetupSuite() {
	s.chainID = int(constant.ChainHSC.ChainID)
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
		initHscHandler,
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

func (s *hscRpcSuite) TearDownSuite() {
	s.httpTestServer.Close()
}

func (s *hscRpcSuite) TearDownTest() {
}

func (s *hscRpcSuite) httpRequest(apikey, body string) []byte {
	httpURL := s.httpTestServer.URL + "/hsc/v1/" + apikey

	req, err := http.NewRequest(http.MethodPost, httpURL, strings.NewReader(body))
	assert.Nil(s.T(), err, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(s.T(), err, err)

	ret, err := ioutil.ReadAll(resp.Body)
	assert.Nil(s.T(), err, err)

	err = resp.Body.Close()
	assert.Nil(s.T(), err, err)

	return ret
}

func (s *hscRpcSuite) createWsConn(apikey string) *websocket.Conn {
	wsURL := "ws" + strings.TrimPrefix(s.httpTestServer.URL, "http") + "/ws/hsc/v1/" + apikey
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.Nil(s.T(), err, err)
	return ws
}

func (s *hscRpcSuite) genAndSetupApikey(secQuota, dayQuota int, t time.Time) string {
FnBegin:
	apikey, err := uuid.GenerateUUID()
	assert.Nil(s.T(), err)

	// Initialize configuration
	err = s.rateLimitDao.SetQuota(apikey, int(s.chainID), secQuota, dayQuota)
	assert.Nil(s.T(), err)

	_, err = s.rateLimitDao.GetDayUsage(apikey, int(s.chainID), t)
	if !errors.Is(err, redis.Nil) {
		// 若key已经存在，则重新生成一个作为测试
		goto FnBegin
	}

	return apikey
}

func (s *hscRpcSuite) newHttpContext(apikey string, body string) (echo.Context, *httptest.ResponseRecorder) {
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

func (s *hscRpcSuite) loadConfig() {
	cfgToml := `
listen = "127.0.0.1:1324"

[upstream]
hsc.http = "https://http-mainnet.hoosmartchain.com"
hsc.ws = "wss://ws-mainnet.hoosmartchain.com"

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
