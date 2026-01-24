package main

import (
	"MyBlog/modules"
	"MyBlog/web/services"
	"fmt"
)

func main() {
	dbm := modules.GetDBManager()
	dbm.Username = "postgres"
	dbm.Password = "lianzeyu29"
	dbm.TcpAddr = "101.37.82.59"
	dbm.TcpPort = "5432"

	err := dbm.Init()
	if err != nil {
		fmt.Println("init db err:", err)
	}
	sm, _ := services.NewServiceManager()
	sm.Start()
}
