package config

import (
	"fmt"

	starnetRedis "starnet/starnet/pkg/redis"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"go.uber.org/zap/buffer"
)

type Config struct {
	Listen string `mapstructure:"listen"`

	Upstream struct {
		Eth struct {
			Http   string `mapstructure:"http"`
			Ws     string `mapstructure:"ws"`
			Erigon struct {
				Http string `mapstructure:"http"`
				Ws   string `mapstructure:"ws"`
			} `mapstructure:"erigon"`
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

	IPFSCluster struct {
		Schemes string `mapstructure:"schemes"`
		Host    string `mapstructure:"host"`
		Port    int    `mapstructure:"port"`
	} `mapstructure:"ipfs_cluster"`

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

type RpcNode struct {
	Name       string `toml:"name"`
	Http       string `toml:"http"`
	Ws         string `toml:"ws"`
	ExtraWrite string `toml:"extra_write"` // tron network only
}

type RpcConfig struct {
	ApiKey            string
	HealthPushgateway string
	Chains            []ChainConfig
}

type ChainConfig struct {
	ChainName                   string
	ChainType                   string    `toml:"chain_type"`
	MaxBehindBlocks             int64     `toml:"max_behind_blocks"`
	BlockNumberMethod           string    `toml:"block_number_method"`
	BlockNumberResultExtractor  string    `toml:"block_number_result_extractor"`
	BlockNumberResultExpression string    `toml:"block_number_result_expression"`
	Nodes                       []RpcNode `toml:"nodes"`
}

func LoadRPCConfig(data string) (*RpcConfig, error) {
	var config map[string]any
	err := toml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	rpcConfig := &RpcConfig{
		ApiKey:            config["apikey"].(string),
		HealthPushgateway: config["health_pushgateway"].(string),
		Chains:            make([]ChainConfig, 0),
	}

	for chainName, chainConfig := range config {
		if chainName == "apikey" || chainName == "health_pushgateway" {
			continue
		}
		var defaultMaxBehindBlocks int64 = 10
		cfg := ChainConfig{
			ChainName: chainName,
			// Default values
			ChainType:                   "evm",
			MaxBehindBlocks:             defaultMaxBehindBlocks,
			BlockNumberMethod:           "",
			BlockNumberResultExtractor:  "jq",
			BlockNumberResultExpression: ".result",
		}
		buf := buffer.Buffer{}
		if err := toml.NewEncoder(&buf).Encode(chainConfig); err != nil {
			return nil, err
		}

		if err := toml.Unmarshal(buf.Bytes(), &cfg); err != nil {
			return nil, err
		}

		switch cfg.ChainType {
		case "evm":
			cfg.BlockNumberMethod = "eth_blockNumber"
		case "svm":
			cfg.BlockNumberMethod = "getBlockHeight"
		case "aptos":
			cfg.BlockNumberResultExpression = ".ledger_version"
		case "tron":
			cfg.BlockNumberMethod = "eth_blockNumber"
			if cfg.MaxBehindBlocks == defaultMaxBehindBlocks {
				cfg.MaxBehindBlocks = 40 // 120 seconds
			}
		default:
			return nil, fmt.Errorf("unsupported chain type: %s", cfg.ChainType)
		}

		rpcConfig.Chains = append(rpcConfig.Chains, cfg)
	}
	return rpcConfig, nil
}
