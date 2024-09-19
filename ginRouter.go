package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// 需要登录才能访问的中间件
func authRequired(mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mode == "debug" {
			log.Println("debug模式不需要登录验证")
			c.Next()
			return
		}
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "未授权", "success": "false"})
			c.Abort()
			return
		}
		log.Printf("通过cookie校验user=%s", user)
		c.Next()
	}
}

func SelectResponseJson[T any](c *gin.Context, datas []*T) {
	var partialDatas []*T
	vol := c.Query("volume") //获取volume参数
	if len(datas) > 200 {
		partialDatas = datas[:200]
	} else {
		partialDatas = datas
	}
	if vol == "all" {
		c.JSON(http.StatusOK, datas) //发送所有数据
	} else if vol == "partial" {
		c.JSON(http.StatusOK, partialDatas) //发送部分数据
	} else {
		c.JSON(http.StatusOK, datas) //发送所有数据
	}
}

// 返回内容为json格式的字符串
func SetGinRouterByJson(r *gin.Engine, mc *MemoryCache) {
	r.GET("/api/login", func(c *gin.Context) {
		log.Println("/login POST require")
		var req BackendUser
		//username := c.PostForm("username") POST请求的话使用PostForm
		req.UserName = c.Query("userName") //获取参数
		req.Password = c.Query("password") //
		var res Response
		var datas []*BackendUser
		mc.GetMemoryCache(&datas, req.UserName, req.Password)
		res.AnyBody = datas
		if len(datas) != 0 {
			res.Success = "true" //校验成功
			log.Printf("name=%s,passwd=%s通过验证", req.UserName, req.Password)
			//如果通过验证需要保存session
			session := sessions.Default(c)
			session.Set("user", req.UserName) // 设置 session
			session.Save()                    // 保存 session
			c.JSON(http.StatusOK, res)
		} else {
			res.Success = "false" //校验失败
			log.Printf("name=%s,passwd=%s验证失败", req.UserName, req.Password)
			c.JSON(http.StatusOK, res)
		}
	})
	r.GET("/api/address", authRequired(mc.Config.Mode), func(c *gin.Context) {
		log.Println("/address GET require")
		var req BackendUser
		var datas []*BackendUser
		req.Address = c.QueryArray("address")
		req.UserName = c.Query("userName")
		datas = append(datas, &req)
		ok := mc.SetMemoryCache(&datas)
		var res Response
		if ok {
			res.Success = "true" //校验成功
		} else {
			res.Success = "false" //校验失败
		}
		c.JSON(http.StatusOK, res)
	})
	r.GET("/api/order", authRequired(mc.Config.Mode), func(c *gin.Context) {
		log.Println("/order GET require")
		var datas []*BackendOrder
		userID := c.Query("userid")
		ok := mc.GetMemoryCache(&datas, userID)
		var res Response
		if ok {
			res.Success = "true" //校验成功
			res.AnyBody = datas
		} else {
			res.Success = "false" //校验失败
		}
		c.JSON(http.StatusOK, res)
	})
	r.POST("/api/order", authRequired(mc.Config.Mode), func(c *gin.Context) {
		rd, _ := c.GetRawData()
		log.Println("/order POST require, rawData = ", string(rd))
		var datas []*BackendOrder
		//ok := ReadOrderPostFormdata(c)
		req := struct {
			op    string          `json:"operation"`
			datas []*BackendOrder `json:"orders"`
		}{}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ok := mc.SetMemoryCache(&req.datas, req.op)
		var res Response
		res.AnyBody = datas
		if ok {
			res.Success = "true" //校验成功
		} else {
			res.Success = "false" //校验失败

		}
		c.JSON(http.StatusOK, res)
	})
}

func ReadOrderPostFormdata(c *gin.Context) bool {
	//读取表单数据
	op := c.PostForm("operation")
	for i := 0; i < 10; i++ {
		id := c.PostForm(fmt.Sprintf("orders[%d][id]", i))
		log.Println("id = ", id)
	}
	log.Println("op = ", op)
	return true
}
