package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	r.GET("/api/address", func(c *gin.Context) {
		//var inventories [](*tablestruct.Inventory)
		log.Println("/address GET require")
		var req BackendUser
		var datas []*BackendUser
		SelectResponseJson(c, datas)
	})
	r.Get("/api/login", func(c *gin.Context) {
		log.Println("/login POST require")
		var req BackendUser
		req.UserName = c.Query("userName") //获取参数
		req.Password = c.Query("password") //
		var res Response
		var datas []*BackendUser
		datas = mc.GetBackendUser(req)
		res.AnyBody = datas
		if len(datas) != 0 {
			res.Success = "true" //校验成功
			log.Printf("name=%s,passwd=%s通过验证", req.UserName, req.Password)
			c.JSON(http.StatusOK, res)
		} else {
			res.Success = "false" //校验失败
			log.Printf("name=%s,passwd=%s验证失败", req.UserName, req.Password)
			c.JSON(http.StatusOK, res)
		}

	})

}
