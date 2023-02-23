package main

import (
	"github.com/Ljkkun/GreenBeanMiners/initialize"
)

func main() {
	initialize.Global() // 初始化全局变量
	initialize.Viper()  // 初始化配置信息
	initialize.MySQL()  // 初始化 MySQL 连接
	initialize.Redis()  // 初始化 Redis 连接
	initialize.Router() // 初始化 GinRouter
	//http.Handle("/public/", http.StripPrefix("/cover/", http.FileServer(http.Dir("./png"))))
	//http.Handle("/public/", http.StripPrefix("/video/", http.FileServer(http.Dir("./mp4"))))
}
