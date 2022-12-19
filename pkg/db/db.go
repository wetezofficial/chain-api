package db

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	gormMySql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
	"starnet/chain-api/config"
	"time"
)

func NewDB(c *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	mysqlConfig := &mysql.Config{
		User:                 c.MySQL.User,
		Passwd:               c.MySQL.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", c.MySQL.Host, c.MySQL.Port),
		DBName:               c.MySQL.Database,
		AllowNativePasswords: true,
	}

	dbLogger := zapgorm2.New(logger)
	dbLogger.SetAsDefault() // optional: configure gorm to use this zapgorm.Logger for callbacks
	if c.Log.Level == "debug" {
		dbLogger.LogLevel = gormlogger.Info
	}

	db, err := gorm.Open(gormMySql.Open(mysqlConfig.FormatDSN()), &gorm.Config{
		Logger:                 dbLogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
