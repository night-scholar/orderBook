package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var matchMsgChan = make(chan OrderParamMsg)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main(){
	orderBook := NewOrderBook()
	r := gin.Default()
	HandlePost(orderBook,r)
	HandleGet(orderBook,r)
	r.GET("/orders", orderBook.Handicap)
	go UpChainBot()
	_ = r.Run(":8080")
}

