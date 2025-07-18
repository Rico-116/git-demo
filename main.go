package main

import (
	"gochatroom/db"
	"gochatroom/server"
	"log"
)

func main() {
	db.InitDB() // 初始化数据库连接（必须在其他操作之前）
	// 启动 TCP 服务器，监听 8888 端口
	err := server.Start(":8888")
	if err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
