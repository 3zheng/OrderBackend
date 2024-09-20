package main

//MemeoyCache Sql与数据库之间的中间件
import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"time"

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
	strsql = "SELECT  `user_name`,  `password`,  `user_id`, `remark_name`, " +
		"`address` " + //最后一个字段是没有逗号的
		" FROM `backend_users` " +
		" where `user_name` = ? and `password` = ?"
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["GetBackendUser"] = stmt
	//更新backend_users
	strsql = "UPDATE `backend_users` " +
		" set `address` = ?, `remark_name` = ? " +
		" where `user_name` = ? "
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["SetBackendUser"] = stmt
	//查询backend_orders
	strsql = "SELECT  `order_id`,  `remark_name`,  `user_id`,  `address`, " +
		"`product_id`, `sub_category`, `product_num`, `order_status`, " +
		"`order_date` " + //最后一个字段是没有逗号的
		" FROM `backend_orders` " +
		" where `user_id` = ? "
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["GetBackendOrder"] = stmt
	//更新backend_orders
	strsql = "UPDATE `backend_orders` set " +
		"`address` = ?, `product_num` = ?, " +
		"`order_status` = ? " + //最后一个字段是没有逗号的
		" where `user_id` = ? and `order_id` = ?"
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["SetBackendOrder"] = stmt
	//新增backend_orders
	strsql = "INSERT backend_orders" +
		"(remark_name,user_id,address, product_id,sub_category, product_num,order_status, order_date)" +
		"VALUES(?,  ?,  ?,  ?,  ?,  ?,  ?, CAST(? AS DATETIME))"
	stmt, err = mc.db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return
	}
	mc.StmtMap["InsertBackendOrder"] = stmt
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
func (mc *MemoryCache) GetMemoryCache(data interface{}, parameters ...string) bool {
	if mc.db == nil {
		log.Println("MemoryCache的数据库实例为空")
		return false
	}
	//先从数据库里取，全部数据都从内存中取
	switch v := data.(type) {
	case *[]*BackendUser:
		if len(parameters) < 2 {
			return false
		}
		var bu BackendUser
		bu.UserName = parameters[0]
		bu.Password = parameters[1]
		*v = mc.GetBackendUser(bu)
	case *[]*BackendOrder:
		if len(parameters) < 1 {
			return false
		}
		var bo BackendOrder
		var err error
		bo.UserID, err = strconv.Atoi(parameters[0])
		if err != nil {
			return false
		}
		*v = mc.GetBackendOrder(bo)
	default:
		log.Println("CheckUpdataCache无法匹配类型:", v)
	}
	return true
}

func (mc *MemoryCache) SetMemoryCache(data interface{}, parameters ...string) bool {
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
	case *[]*BackendOrder:
		if len(parameters) < 1 {
			return false
		}
		op := parameters[0]
		var ok bool
		for _, bo := range *v {
			switch op {
			case "update":
				ok = mc.SetBackendOrder(*bo)
			case "insert":
				ok = mc.InsertBackendOrder(*bo)
			case "delete":
				ok = mc.DeleteBackendOrder(*bo)
			default:
				log.Println("未知的BackendOrder操作类型：", op)
			}
			if !ok {
				log.Printf("BackendOrder failed: op=%s, %#v", op, *bo)
			}
		}
		return ok
	default:
		log.Println("CheckUpdataCache无法匹配类型:", v)
		return false
	}
	return true
}

type BackendUser struct {
	//select 客户ID,客户姓名,月开始日期,月购买总金额,月购买次数 from NewKeyCustomers
	UserID     int      `json:"ID"`         //ID
	UserName   string   `json:"UserName"`   //用户名
	Password   string   `json:"Password"`   //密码
	Address    []string `json:"Address"`    //地址
	RemarkName string   `json:"RemarkName"` //备注名
}

func (mc *MemoryCache) GetBackendUser(param BackendUser) [](*BackendUser) {
	//返回的数据集
	var rowsData [](*BackendUser)
	stmt, ok := mc.StmtMap["GetBackendUser"]
	if !ok {
		log.Println("stmt not found: GetBackendUser")
		return rowsData
	}
	//执行查询语句"SELECT  `user_name`,  `password`,  `user_id`, `remark_name`, `address` FROM `backend_user` where `user_name` = ? and `password` = ?"
	//传入参数是user_name和password
	rows, err := stmt.Query(param.UserName, param.Password)
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
			&data.UserID, &data.RemarkName, &jsonString)
		// 将 JSON 字符串反序列化为结构体
		err := json.Unmarshal(jsonString, &data.Address)
		if err != nil {
			log.Println("JSON反序列化错误")
		}
		rowsData = append(rowsData, data)
	}
	return rowsData
}

func (mc *MemoryCache) SetBackendUser(param BackendUser) bool {
	//返回的数据集
	stmt, ok := mc.StmtMap["SetBackendUser"]
	if !ok {
		log.Println("找不到")
		return false
	}

	//把[]string转为json格式
	jsonData, err := json.Marshal(param.Address)
	if err != nil {
		log.Panicln("Address序列化JSON失败:", param.Address)
		return false
	}
	// 执行SQL语句"update `backend_user` set `address` = ?, remark_name = ?  where `user_name` = ? "
	//传递address和user_name参数
	result, err := stmt.Exec(jsonData, param.RemarkName, param.UserName)
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
		jsonData, param.UserName, rowsAffected)
	return true
}

type BackendOrder struct {
	//select 客户ID,客户姓名,月开始日期,月购买总金额,月购买次数 from NewKeyCustomers
	OrderID     int    `json:"OrderID"`     //订单号
	RemarkName  string `json:"UserName"`    //用户名
	UserID      int    `json:"UserID"`      //用户ID
	Address     string `json:"Address"`     //地址
	ProductID   string `json:"ProductID"`   //产品ID，是字符串类型
	SubCategory string `json:"SubCategory"` //子类
	ProductNum  int    `json:"ProductNum"`  //订购数量
	Status      int    `json:"Status"`      //订单状态
	OrderDate   string `json:"OrderDate"`   //订单日期
}

func (mc *MemoryCache) GetBackendOrder(param BackendOrder) [](*BackendOrder) {
	//返回的数据集
	var rowsData [](*BackendOrder)
	stmt, ok := mc.StmtMap["GetBackendUser"]
	if !ok {
		log.Println("stmt not found: GetBackendUser")
		return rowsData
	}
	// 执行查询语句 "SELECT  `order_id`,  `remark_name`,  `user_id`,  `address`, `product_id`, `sub_category`, `product_num`, `order_status`, `order_date` where `user_id` = ? "
	// 传入参数是 user_id
	rows, err := stmt.Query(param.UserID)
	if err != nil {
		log.Println("Query failed:", err.Error())
		return nil
	}
	//将数据读取到实体中
	for rows.Next() {
		data := new(BackendOrder)
		var dt time.Time
		//其中一个字段的信息 ， 如果要获取更多，就在后面增加：rows.Scan(&row.Name,&row.Id)
		rows.Scan(&data.OrderID, &data.RemarkName, &data.UserID, &data.Address,
			&data.ProductID, &data.SubCategory, &data.ProductNum, &data.Status,
			&dt)
		data.OrderDate = dt.Format(time.DateTime)
		// 将 JSON 字符串反序列化为字符串数组
		rowsData = append(rowsData, data)
	}
	return rowsData
}

func (mc *MemoryCache) SetBackendOrder(param BackendOrder) bool {
	//返回的数据集
	stmt, ok := mc.StmtMap["SetBackendOrder"]
	if !ok {
		log.Println("找不到")
		return false
	}

	// 执行SQL语句"UPDATE `backend_orders` set `address` = ?, `product_num` = ?, `order_status` = ?  where `user_id` = ? and `order_id` = ?"
	//传递address, product_num, status, user_id, order_id参数
	result, err := stmt.Exec(param.Address, param.ProductNum, param.Status, param.UserID, param.OrderID)
	if err != nil {
		log.Println("Error executing statement:", err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return false
	}

	log.Printf("SetBaBackendOrder Address=%s, ProductNum=%d, Status=%d, UserID=%d, OrderID=%d Rows affected: %d\n",
		param.Address, param.ProductNum, param.Status, param.UserID, param.OrderID, rowsAffected)
	return true
}

func (mc *MemoryCache) InsertBackendOrder(param BackendOrder) bool {
	//返回的数据集
	stmt, ok := mc.StmtMap["InsertBackendOrder"]
	if !ok {
		log.Println("InsertBackendOrder找不到stmt")
		return false
	}

	//执行SQL语句strsql = "INSERT backend_orders (user_name,user_id,address, product_id,sub_category, product_num, order_status, order_date) VALUES(?,  ?,  ?,  ?,  ?,  ?,  ?, CAST(? AS DATETIME))"
	//传递 user_name, user_id, address, product_id, sub_category, product_num,order_status,order_date 参数
	result, err := stmt.Exec(param.RemarkName, param.UserID, param.Address, param.ProductID,
		param.SubCategory, param.ProductNum, 0, param.OrderDate) //order_status为0，新订单的状态都是0
	if err != nil {
		log.Println("Error executing statement:", err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return false
	}

	log.Printf("InsertBackendOrder Address=%s, ProductNum=%d, Status=%d, UserID=%d, OrderID=%d Rows affected: %d\n",
		param.Address, param.ProductNum, param.Status, param.UserID, param.OrderID, rowsAffected)
	return true
}

func (mc *MemoryCache) DeleteBackendOrder(param BackendOrder) bool {
	return true
}
