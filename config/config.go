package config

import (
	"github.com/jinzhu/configor"
	"github.com/liyue201/go-logger"
)

type (
	BtcNodeCfg struct {
		Network     string `json:"network" env:"BTCNODE_NETWORK"`
		Certificate string `json:"certificate" env:"BTCNODE_CERTFICATE"`
		Address     string `json:"address" env:"BTCNODE_ADDRESS"`
		User        string `json:"user" env:"BTCNODE_USER"`
		Password    string `json:"password" env:"BTCNODE_PASSWORD"`
		Mempool     int    `json:"mempool" env:"BTCNODE_MEMPOOL"`
	}

	SSDBConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}

	RestConfig struct {
		Port int `json:"port"  env:"REST_PORT"`
	}

	LevelDBConfig struct {
		Path string `json:"path" env:"LEVELDB_PATH"`
	}

	Configuration struct {
		Name      string        `json:"name"`
		Btcnode   BtcNodeCfg    `json:"btcnode"`
		Ssdb      SSDBConfig    `json:"ssdb"`
		Leveldb   LevelDBConfig `json:"leveldb"`
		Rest      RestConfig    `json:"rest"`
		Cachesize int           `json:"cachesize" env:"CACHESIZE"`
	}
)

var Cfg Configuration

func Init(filePath string) error {
	err := configor.Load(&Cfg, filePath)
	logger.Infof("config: %#v", Cfg)
	return err
}
