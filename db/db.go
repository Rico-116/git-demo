package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	dsn := "root:YRK1314k&@tcp(127.0.0.1:3306)/chatdb?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	fmt.Println("数据库连接成功")
}
