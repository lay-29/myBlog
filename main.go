package main

import (
	"MyBlog/modules"
	"MyBlog/web/services"
	"fmt"
)

func main() {
	dbm := modules.DBManager{Username: "postgres", Password: "lianzeyu29", TcpAddr: "101.37.82.59", TcpPort: "5432"}
	err := dbm.Init()
	if err != nil {
		fmt.Println("init db err:", err)
	}
	sm, _ := services.NewServiceManager()
	sm.Start()
}
