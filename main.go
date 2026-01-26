package main

import (
	"MyBlog/models"
	"MyBlog/modules"
	"MyBlog/web/services"
	"fmt"
)

func main() {
	appCfg := models.GetAppConfig()
	if err := appCfg.LoadOrCreateConfig(".\\config.ini"); err != nil {
		fmt.Println("init appCfg err:", err)
	}
	dbm := modules.GetDBManager()
	dbm.Username = models.GetAppConfig().DB.Username
	dbm.Password = models.GetAppConfig().DB.Password
	dbm.TcpAddr = models.GetAppConfig().DB.TcpAddr
	dbm.TcpPort = models.GetAppConfig().DB.TcpPort

	err := dbm.Init()
	if err != nil {
		fmt.Println("init db err:", err)
	}
	sm, _ := services.NewServiceManager()
	sm.Start()
}
