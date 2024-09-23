package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	util "github.com/3zheng/go_util"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	defer util.Recovermain()
	cfg := util.ReadConfigFile()
	go util.InitLog(cfg)

	//mysql的连接字符串格式
	//注意charset应该是utf8mb4而不是utf8mb4_general_ci,前者是字符集，后者是排序规则
	//connString := "username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4_general_ci"

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		cfg.Database.UserId, cfg.Database.Password, cfg.Database.IP, cfg.Database.Port, cfg.Database.DB)

	//建立SQLSever数据库连接：db
	db, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal("Open Connection failed:", err.Error())
	}
	log.Println("建立数据库连接")
	defer db.Close()
	mc := new(MemoryCache)
	mc.InitMemoryCache(db, cfg)
	defer mc.ClosePrepare()
	//启动gin
	r := gin.Default()
	r.Use(cors.Default()) //使用cors，解决跨域问题
	if cfg.Server.CookieKey == "" {
		cfg.Server.CookieKey = "secret-key"
	}
	store := cookie.NewStore([]byte(cfg.Server.CookieKey)) //根据配置的密钥初始化cookie
	r.Use(sessions.Sessions("mysession", store))
	//
	SetGinRouterByJson(r, mc) //返回json数据，前端后端分离，后端只返回数据，前端不管

	log.Println("开始启动web服务")
	addr := fmt.Sprintf("%s:%d", cfg.Server.IP, cfg.Server.Port)
	//ln := net.Listener
	if cfg.Server.ForceIPv4 == 1 {
		// 强制使用IPv4
		log.Println("强制使用IPv4")
		server := &http.Server{Addr: addr, Handler: r}
		ln, err := net.Listen("tcp4", addr)
		if err != nil {
			panic(err)
		}
		type tcpKeepAliveListener struct {
			*net.TCPListener
		}

		server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
	} else {
		log.Println("http服务地址：", addr)
		r.Run(addr)
	}
}
