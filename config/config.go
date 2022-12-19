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
		Kava struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"kava"`
		Juno struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"juno"`
		Umee struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"umee"`
		Gravity struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"gravity"`
		OKC struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"okc"`
		IRISnet struct {
			Http string `mapstructure:"http"`
			Ws   string `mapstructure:"ws"`
		} `mapstructure:"irisnet"`
	} `mapstructure:"upstream"`

	Log struct {
		Level         string `mapstructure:"level"`
		IsDevelopment bool   `mapstructure:"is_dev"`
		LogFile       string `mapstructure:"log_file"`
	} `mapstructure:"log"`

	MySQL struct {
		Database string `mapstructure:"database"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	} `mapstructure:"mysql"`

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
		return nil, fmt.Errorf("读取配置文件：%w", err)
	}

	return cfg, nil
}
