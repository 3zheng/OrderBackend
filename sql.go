package main

import (
	"database/sql"
	"log"
	"time"

	util "github.com/3zheng/go_util"
)

type MemoryCache struct {
	db      *sql.DB     //数据库实例
	Config  util.Config //配置文件
	StmtMap map[string]*sql.Stmt
}

func (mc *MemoryCache) PrepareDBSql() {
	if mc.db == nil {
		panic("数据库为空")
		return
	}

}

func (mc *MemoryCache) ClosePrepare() {

}

type Wordpress struct {
	//select 客户ID,客户姓名,月开始日期,月购买总金额,月购买次数 from NewKeyCustomers
	ID             int       `json:"ID"`             //客户ID
	Post_author    string    `json:"Post_author"`    //客户姓名
	Post_date      time.Time `json:"Post_date"`      //月开始日期
	Post_date_gmt  time.Time `json:"Post_date_gmt"`  //月开始日期
	Post_content   string    `json:"Post_content"`   //客户姓名
	Post_title     string    `json:"Post_title"`     //客户姓名
	Post_status    string    `json:"Post_status"`    //客户姓名
	Comment_status string    `json:"Comment_status"` //客户姓名
	Ping_status    string    `json:"Ping_status"`    //客户姓名
	Post_password  string    `json:"Post_password"`  //客户姓名
	Post_name      string    `json:"Post_name"`      //客户姓名
	Post_parent    int       `json:"Post_parent"`    //客户ID
	Guid           string    `json:"Guid"`           //客户姓名
	Menu_order     int       `json:"Menu_order"`     //客户ID
	Post_type      string    `json:"Post_type"`      //客户姓名
	Comment_count  int       `json:"Comment_count"`  //客户ID
}

func GetWordpress(db *sql.DB) [](*Wordpress) {
	//编写查询语句
	//select 客户ID,客户姓名,月开始日期,月购买总金额,月购买次数 from dbo.CustomerYearlySalesReport
	strsql := "SELECT  `ID`,  `post_author`,  `post_date`,  `post_date_gmt`,  " +
		"LEFT(`post_content`, 256), LEFT(`post_title`, 256),  " +
		"`post_status`,  `comment_status`,  `ping_status`,  `post_password`,  `post_name`,  " +
		"`post_parent`,  `guid`,  `menu_order`, `post_type`, `comment_count`" +
		" FROM `wordpress`.`wp_posts`;"
	stmt, err := db.Prepare(strsql)
	if err != nil {
		log.Println("Prepare failed:", err.Error())
		return nil
	}
	defer stmt.Close()
	//执行查询语句
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Query failed:", err.Error())
		return nil
	}
	//将数据读取到实体中
	var rowsData [](*Wordpress)
	for rows.Next() {
		data := new(Wordpress)

		//其中一个字段的信息 ， 如果要获取更多，就在后面增加：rows.Scan(&row.Name,&row.Id)
		rows.Scan(&data.ID, &data.Post_author, &data.Post_date, &data.Post_date_gmt,
			&data.Post_content, &data.Post_title, &data.Post_status, &data.Comment_status,
			&data.Ping_status, &data.Post_password, &data.Post_name,
			&data.Post_parent, &data.Guid, &data.Menu_order,
			&data.Post_type, &data.Comment_count)

		rowsData = append(rowsData, data)
	}
	return rowsData
}
