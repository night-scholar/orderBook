package main

import (
	"github.com/gin-gonic/gin"
)

func main(){
	orderBook := NewOrderBook()
	r := gin.Default()
	HandlePost(orderBook,r)
	HandleGet(orderBook,r)
	_ = r.Run(":8838")
}

