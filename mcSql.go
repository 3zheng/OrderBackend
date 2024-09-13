package main

//MemeoyCache Sql与数据库之间的中间件
import (
	"database/sql"
	"encoding/json"
	"log"

	util "github.com/3zheng/go_util"
)

type MemoryCache struct {
	db      *sql.DB     //数据库实例
	Config  util.Config //配置文件
	StmtMap map[string]*sql.Stmt
}

type Response struct {
	Success string      `json:"Success"` //0：失败， 1：成功
	AnyBody interface{} `json:"AnyBody"` //返回的数据
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
	strsql = "SELECT  `user_name`,  `password`,  `user_id`,  `address` " +
		" FROM `backend_users` " +
		" where `user_name` = ? and `password` = ?"
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["GetBackendUser"] = stmt
	//查询
	strsql = "update `backend_users` " +
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

// 判断是否需要更新缓存，如果数据超过了一分钟，则更新，否则不更新
func (mc *MemoryCache) GetMemoryCache(data interface{}, parameters ...string) {
	if mc.db == nil {
		log.Println("MemoryCache的数据库实例为空")
		return
	}
	//先从数据库里取，全部数据都从内存中取
	switch v := data.(type) {
	case *[]*BackendUser:
		if len(parameters) < 2 {
			return
		}
		var bu BackendUser
		bu.UserName = parameters[0]
		bu.Password = parameters[1]
		*v = mc.GetBackendUser(bu)
	default:
		log.Println("CheckUpdataCache无法匹配类型:", v)
	}
}

func (mc *MemoryCache) UpdateDataBase(data interface{}, parameters ...string) bool {
	if mc.db == nil {
		log.Println("MemoryCache的数据库实例为空")
		return false
	}
	//先从数据库里取，全部数据都从内存中取
	switch v := data.(type) {
	case *[]*BackendUser:
		//更新输入的[]BackendUser,不过通常只会有一个元素
		for _, bu := range *v {
			if !mc.SetBackendUser(*bu) {
				var address string
				for _, add := range bu.Address {
					address += add
				}
				log.Printf("SetBackendUser failed: username=%s, address=%s", bu.UserName, address)
				return false
			}
		}
	default:
		log.Println("CheckUpdataCache无法匹配类型:", v)
		return false
	}
	return true
}

type BackendUser struct {
	//select 客户ID,客户姓名,月开始日期,月购买总金额,月购买次数 from NewKeyCustomers
	UserID   int      `json:"ID"`       //客户ID
	UserName string   `json:"UserName"` //客户姓名
	Password string   `json:"Password"` //客户姓名
	Address  []string `json:"Address"`  //客户姓名
}

func (mc *MemoryCache) GetBackendUser(user BackendUser) [](*BackendUser) {
	//返回的数据集
	var rowsData [](*BackendUser)
	stmt, ok := mc.StmtMap["GetBackendUser"]
	if !ok {
		log.Println("stmt not found: GetBackendUser")
		return rowsData
	}
	//执行查询语句"SELECT  `user_name`,  `password`,  `user_id`,  `address`,  FROM `backend_user` where `user_name` = ? and `password` = ?"
	//传入参数是user_name和password
	rows, err := stmt.Query(user.UserName, user.Password)
	if err != nil {
		log.Println("Query failed:", err.Error())
		return nil
	}
	//将数据读取到实体中
	for rows.Next() {
		data := new(BackendUser)
		var jsonString []byte
		//其中一个字段的信息 ， 如果要获取更多，就在后面增加：rows.Scan(&row.Name,&row.Id)
		rows.Scan(&data.UserName, &data.Password,
			&data.UserID, &jsonString)
		// 将 JSON 字符串反序列化为字符串数组
		err := json.Unmarshal(jsonString, &data.Address)
		if err != nil {
			log.Println("JSON反序列化错误")
		}
		rowsData = append(rowsData, data)
	}
	return rowsData
}

func (mc *MemoryCache) SetBackendUser(user BackendUser) bool {
	//返回的数据集
	stmt, ok := mc.StmtMap["SetBackendUser"]
	if !ok {
		log.Println("找不到")
		return false
	}

	//把[]string转为json格式
	jsonData, err := json.Marshal(user.Address)
	if err != nil {
		panic(err)
	}
	// 执行SQL语句"update `backend_user` set `address` = ?  where `user_name` = ? "
	//传递address和user_name参数
	result, err := stmt.Exec(jsonData, user.UserName)
	if err != nil {
		log.Println("Error executing statement:", err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return false
	}

	log.Printf("SetBackendUser address=%s, user_name=%s Rows affected: %d\n",
		jsonData, user.UserName, rowsAffected)
	return true
}
