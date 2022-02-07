package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net"
	"sync"
	"time"
)

type Handicap struct {
	BidsHandicap []*OrderQueue `json:"bids"`
	AsksHandicap []*OrderQueue `json:"asks"`
}

type HandicapSocket struct {
	Type string `json:"type"`
	Data *Handicap `json:"data"`
}

type DealsSocket struct {
	Type string `json:"type"`
	Data []*HistoryDeal `json:"data"`
}

func NewHandicap() *Handicap{
	return &Handicap{
		BidsHandicap: []*OrderQueue{},
		AsksHandicap: []*OrderQueue{},
	}
}

//webSocket请求盘口
func (ob *OrderBook)Handicap(c *gin.Context)  {
	//升级get请求为websocket协议
	ws,err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	mutex := sync.Mutex{}
	go func() {
		for {
			AskSlice := make([]*OrderQueue,0)
			BidSlice := make([]*OrderQueue,0)
			if len(ob.handicap.AsksHandicap) < 30{
				for i := 0 ; i < 30 - len(ob.handicap.AsksHandicap) ;i ++{
					AskSlice = append(AskSlice, NewOrderQueue(decimal.Zero))
				}
				AskSlice = append(AskSlice, ob.handicap.AsksHandicap...)
			}else{
				AskSlice = append(AskSlice, ob.handicap.AsksHandicap[len(ob.handicap.AsksHandicap)-30 : len(ob.handicap.AsksHandicap)]...)
			}

			if len(ob.handicap.BidsHandicap) < 30{
				BidSlice = append(BidSlice, ob.handicap.BidsHandicap[0:len(ob.handicap.BidsHandicap)]...)
				for i:= 0 ;i < 30 -len(ob.handicap.BidsHandicap) ;i++{
					BidSlice = append(BidSlice, NewOrderQueue(decimal.Zero))
				}
			}else{
				BidSlice = append(BidSlice , ob.handicap.BidsHandicap[0:30]...)
			}
			handicapData := NewHandicap()
			handicapData.BidsHandicap = BidSlice
			handicapData.AsksHandicap = AskSlice
			handicap := HandicapSocket{
				Type: "handicap",
				Data: handicapData,
			}

			buf ,err := json.Marshal(handicap)
			mutex.Lock()
			err = ws.WriteMessage(1,buf)
			mutex.Unlock()
			time.Sleep(3*time.Second)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	}()


	for {
		//mt, message, err := ws.ReadMessage()
		//if err != nil {
		//	break
		//}
		//fmt.Println(string(message))
		var newHistoryDeal = make([]*HistoryDeal ,0)
		for i := len(ob.historyDeal)-1 ; i >= 0  ; i--{
			newHistoryDeal = append(newHistoryDeal, ob.historyDeal[i])
		}
		deals := DealsSocket{
			Type: "deals",
			Data: newHistoryDeal,
		}

		buf ,err := json.Marshal(deals)

		mutex.Lock()
		err = ws.WriteMessage(1,buf)
		mutex.Unlock()
		time.Sleep(10*time.Second)
		if err != nil {
			break
		}
	}

}

//上链机器人请求
func UpChainBot() {
	//主动连接服务器
	conn,err:=net.Dial("tcp","192.168.0.139:2526")
	if err!=nil {
		fmt.Println("连接服务器err = ",err)
		return
	}
	defer conn.Close()


	//发送数据
	for  {
		message := <- matchMsgChan
		buf ,err := json.Marshal(message)
		_,err =conn.Write(buf)
		if err!=nil {
			fmt.Println("发送数据err = ",err)
			return
		}
	}
}
