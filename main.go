package main

import (
	"database/sql"
	"fmt"
	"log"

	util "github.com/3zheng/go_util"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	defer util.Recovermain()
	cfg := util.ReadConfigFile()
	go util.InitLog(cfg)

	//mysql的连接字符串格式
	//connString := "username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4_general_ci"

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4_general_ci",
		cfg.Database.UserId, cfg.Database.Password, cfg.Database.IP, cfg.Database.Port, cfg.Database.DB)

	//建立SQLSever数据库连接：db

	db, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal("Open Connection failed:", err.Error())
	}
	log.Println("建立数据库连接")
	defer db.Close()
}
