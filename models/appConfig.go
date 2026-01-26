package models

import (
	"gopkg.in/ini.v1"
	"os"
	"sync"
)

var (
	instance *AppConfig
	once     sync.Once
)

type AppConfig struct {
	DB *DBSection `ini:"db"`
}

type DBSection struct {
	TcpAddr  string
	TcpPort  string
	Username string
	Password string
}

func GetAppConfig() *AppConfig {
	once.Do(func() {
		instance = new(AppConfig)
		instance.DB = &DBSection{TcpAddr: "123", TcpPort: "3306", Username: "123", Password: "123"}
	})
	return instance
}
func (appCfg *AppConfig) LoadOrCreateConfig(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		iniFile := ini.Empty()
		_ = iniFile.ReflectFrom(appCfg)
		err = iniFile.SaveTo(path)
		return err
	}
	iniFile, err := ini.Load(path)
	if err != nil {
		return err
	}
	err = iniFile.MapTo(appCfg)
	return err
}
