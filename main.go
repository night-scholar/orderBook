package main

import (
	"github.com/gin-gonic/gin"
)

//项目初始化
func init(){
	//读取数据库订单信息
	//append到订单簿
}

func main(){
	orderBook := NewOrderBook()
	r := gin.Default()
	HandlePost(orderBook,r)
	HandleGet(orderBook,r)
	_ = r.Run(":8838")
}