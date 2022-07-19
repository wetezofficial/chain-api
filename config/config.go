package config

import (
	"fmt"

	starnetRedis "starnet/starnet/pkg/redis"

	"github.com/spf13/viper"
)

type Config struct {
	Listen string `mapstructure:"listen"`

	Upstream struct {
		Eth struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"eth"`
		Polygon struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"polygon"`
		Arbitrum struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"arbitrum"`
		Solana struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"solana"`
		Hsc struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"hsc"`
		Cosmos struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"cosmos"`
		Evmos struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"evmos"`
		Gravity struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"Gravity"`
	} `mapstructure:"upstream"`

	Log struct {
		Level         string `mapstructure:"level"`
		IsDevelopment bool   `mapstructure:"is_dev"`
		LogFile       string `mapstructure:"log_file"`
	} `mapstructure:"log"`

	Redis []starnetRedis.Conf `mapstructure:"redis"`
}

func LoadConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("ReadInConfig: %w", err)
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件: %w", err)
	}

	return cfg, nil
}
