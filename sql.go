package main

import (
	"database/sql"
	"log"

	util "github.com/3zheng/go_util"
)

type MemoryCache struct {
	db      *sql.DB     //数据库实例
	Config  util.Config //配置文件
	StmtMap map[string]*sql.Stmt
}

type Response struct {
	Success string      //0：失败， 1：成功
	AnyBody interface{} //返回的数据
}

func (mc *MemoryCache) InitMemoryCache(db *sql.DB, config util.Config) {
	if db == nil {
		log.Println("MemoryCache的数据库实例为空")
		return
	}
	mc.StmtMap = make(map[string]*sql.Stmt)
	mc.db = db
	mc.Config = config
	mc.PrepareDBSql() //预编译
}

// 预编译SQL增加执行效率
func (mc *MemoryCache) PrepareDBSql() {
	if mc.db == nil {
		panic("数据库为空")
	}
	var stmt *sql.Stmt
	var err error
	var strsql string
	//编写查询语句，取后台用户，校验密码
	strsql = "SELECT  `user_name`,  `password`,  `user_id`,  `address`,  " +
		" FROM `backend_user` " +
		" where `user_name` = ? and `password` = ?"
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["GetBackendUser"] = stmt
	//查询
	strsql = "update backend_user" +
		" set `address` = ? " +
		" where `user_name` = ? "
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["SetBackendUser"] = stmt
}

// 关闭预编译SQL
func (mc *MemoryCache) ClosePrepare() {
	for _, v := range mc.StmtMap {
		if v == nil {
			continue
		}
		v.Close()
	}
}

type BackendUser struct {
	//select 客户ID,客户姓名,月开始日期,月购买总金额,月购买次数 from NewKeyCustomers
	ID       int    `json:"ID"`       //客户ID
	UserName string `json:"UserName"` //客户姓名
	Password string `json:"Password"` //客户姓名
	Address  []byte `json:"Address"`  //客户姓名
}

func (mc *MemoryCache) GetBackendUser(user BackendUser) [](*BackendUser) {
	//返回的数据集
	var rowsData [](*BackendUser)
	stmt, ok := mc.StmtMap["GetBackendUser"]
	if !ok {
		log.Println("找不到")
		return rowsData
	}
	//执行查询语句
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Query failed:", err.Error())
		return nil
	}
	//将数据读取到实体中
	for rows.Next() {
		data := new(BackendUser)
		//其中一个字段的信息 ， 如果要获取更多，就在后面增加：rows.Scan(&row.Name,&row.Id)
		rows.Scan(&data.ID, &data.UserName,
			&data.Password, &data.Address,
		)
		rowsData = append(rowsData, data)
	}
	return rowsData
}

func (mc *MemoryCache) SetBackendUser(userName string, password string) [](*BackendUser) {
	//返回的数据集
	var rowsData [](*BackendUser)
	stmt, ok := mc.StmtMap["GetBackendUser"]
	if !ok {
		log.Println("找不到")
		return rowsData
	}
	//执行查询语句
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Query failed:", err.Error())
		return nil
	}
	//将数据读取到实体中
	for rows.Next() {
		data := new(BackendUser)
		//其中一个字段的信息 ， 如果要获取更多，就在后面增加：rows.Scan(&row.Name,&row.Id)
		rows.Scan(&data.ID, &data.UserName,
			&data.Password, &data.Address,
		)
		rowsData = append(rowsData, data)
	}
	return rowsData
}
