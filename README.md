# OrderBackend

## 启动项目需要新建一个config.json, 文件内容如下
```
{
    "project name":"Order Backend",
    "database config":{
        "ip":"127.0.0.1",
        "port":3306,
        "database":"dbname",
        "user id":"root",
        "password":"password"
    },
    "server config":{
        "path":"/home/user/workspace",
        "force ipv4":0,
        "ip":"",
        "port":8889
    },
	"MysqlConn":"root:wzblvroot+2s@tcp(127.0.0.1:3306)/wordpress?charset=utf8mb4_general_ci",

    "none":"没有意义的占位行,用于每行复制逗号用,简单注释也放可以放这里。\
            database config.user id是数据库的用户名, \
            server config.force ipv4表示是否强制使用ipv4监听"
}
```



