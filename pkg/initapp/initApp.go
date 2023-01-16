package initapp

import (
	"fmt"
	"log"

	"starnet/chain-api/pkg/db"
	"starnet/chain-api/service"
	"starnet/portal-api/pkg/cache"

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

	logger := config.NewLogger(cfg)

	rdb := starnetRedis.New(cfg.Redis)
	var rateLimitDao daoInterface.RateLimitDao = dao.NewRateLimitDao(rdb)
	apiKeysWhitelist, err := rateLimitDao.GetApiKeysWhitelist()
	if err != nil {
		log.Fatalf("获取apikey白名单出错: %s\n", err.Error())
	}
	fmt.Println("api keys whitelist:", apiKeysWhitelist)

	_db, err := db.NewDB(cfg, logger)
	if err != nil {
		log.Fatalln(err)
	}

	ipfsDao := dao.NewIPFSDao(_db)
	userDao := dao.NewUserDao(_db)

	rdbCache := cache.NewRedisCache(rdb, "chain:")

	ipfsSrv := service.NewIpfsService(ipfsDao, userDao, rdbCache, logger)

	rateLimiter, err := ratelimitv1.NewRateLimiter(rdb, ipfsSrv, logger, apiKeysWhitelist)
	if err != nil {
		logger.Fatal("fail to get rate limiter", zap.Error(err))
	}

	_app := app.App{
		Config:      cfg,
		Logger:      logger,
		Rdb:         rdb,
		DB:          _db,
		RateLimiter: rateLimiter,
		IPFSSrv:     ipfsSrv,
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
		initIRISnetHandler,
		initIPFSClient,
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
