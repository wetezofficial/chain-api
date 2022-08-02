package initapp

import (
	"fmt"
	"log"

	"starnet/chain-api/config"
	"starnet/chain-api/pkg/app"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"
	"starnet/chain-api/router"
	"starnet/starnet/dao"
	daoInterface "starnet/starnet/dao/interface"
	starnetRedis "starnet/starnet/pkg/redis"

	"go.uber.org/zap"
)

func NewApp(configFile string) *app.App {
	// load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("LoadConfig: %s \n", err.Error())
	}

	logger, err := config.NewLogger(cfg)
	if err != nil {
		log.Fatalf("NewLogger: %s \n", err.Error())
	}

	rdb := starnetRedis.New(cfg.Redis)

	var rateLimitDao daoInterface.RateLimitDao = dao.NewRateLimitDao(rdb)
	apiKeysWhitelist, err := rateLimitDao.GetApiKeysWhitelist()
	if err != nil {
		log.Fatalf("获取apikey白名单出错: %s\n", err.Error())
	}
	fmt.Println("api keys whitelist:", apiKeysWhitelist)

	rateLimiter, err := ratelimitv1.NewRateLimiter(rdb, logger, apiKeysWhitelist)
	if err != nil {
		logger.Fatal("fail to get rate limiter", zap.Error(err))
	}

	_app := app.App{
		Config:      cfg,
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
		initCosmosHandler,
		initEvmosHandler,
		initKavaHandler,
		initJunoHandler,
		initUmeeHandler,
		initGravityHandler,
		initOKCHandler,
	}

	for _, fn := range initFns {
		if err = fn(&_app); err != nil {
			logger.Fatal("fail to run init func", zap.Error(err))
		}
	}

	_router := router.NewRouter(&_app)
	_app.HttpServer = _router

	return &_app
}
