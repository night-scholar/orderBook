package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"sort"
	"strconv"
)

func HandleGet(orderBook *OrderBook,r *gin.Engine)  {
	r.GET("/getOrder" , func(c *gin.Context) {
		types := c.Query("type")
		if types == "all"{
			var bidsHandicap []*OrderQueue
			var asksHandicap []*OrderQueue
			var bidSlice []float64
			var askSlice []float64
			for k,_ := range orderBook.bids.prices{
				key,_ := strconv.ParseFloat(k,64)
				bidSlice = append(bidSlice,key)
			}
			sort.Float64s(bidSlice)
			for k,_ := range orderBook.asks.prices{
				key,_ := strconv.ParseFloat(k,64)
				askSlice = append(askSlice,key)
			}
			sort.Float64s(askSlice)

			if bidSlice != nil{
				for i := len(orderBook.bids.prices)-1 ;i>=0 ; i-- {
					bidsHandicap = append(bidsHandicap, orderBook.bids.prices[strconv.FormatFloat(bidSlice[i],'f',-1,64)])
				}
			}
			if askSlice != nil{
				for i := len(orderBook.asks.prices)-1 ;i>=0 ; i-- {
					asksHandicap = append(asksHandicap, orderBook.asks.prices[strconv.FormatFloat(askSlice[i],'f',-1,64)])
				}
			}

			c.JSON(http.StatusOK , gin.H{
				"code": http.StatusOK,
				"msg": "ok",
				"data":gin.H{
					"bids" :bidsHandicap,
					"asks" :asksHandicap,
				},
			})
		}else if types == "deals"{
			c.JSON(http.StatusOK , gin.H{
				"code": http.StatusOK,
				"msg": "ok",
				"data":orderBook.historyDeal,
			})
		}else if types == "position"{
			trader := c.Query("trader")
			if trader != ""{
				c.JSON(http.StatusOK , gin.H{
					"code": http.StatusOK,
					"msg": "ok",
					"data": orderBook.traderDeal[trader],
				})
			}else{
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusBadRequest,
					"msg": "trader can not be null",
				})
			}
		}else if types == "depth"{
			c.JSON(http.StatusOK , gin.H{
				"code": http.StatusOK,
				"msg": "ok",
				"data": gin.H{
					"askDepth" :orderBook.asks.Depth(),
					"bidDepth" :orderBook.bids.Depth(),
				},
			})
		}else if types == "traderOrder"{
			trader := c.Query("trader")
			if trader != ""{
				c.JSON(http.StatusOK,gin.H{
					"code": http.StatusOK,
					"msg": "ok",
					"data":orderBook.OrderByAddress(trader),
				})
			}else{
				c.JSON(http.StatusOK,gin.H{
					"code":http.StatusBadRequest,
					"msg": "trader can not be Null",
				})
			}
		}else {
			c.JSON(http.StatusOK , gin.H{
				"code":http.StatusBadRequest,
				"msg": "type error",
			})
		}
	})
}

type OrderParam struct {
	Trader string `json:"trader"`
	Amount string `json:"amount"`
	Price string `json:"price"`
	Data string `json:"data"`
	Signature string `json:"signature"`
}

func HandlePost(orderBook *OrderBook,r *gin.Engine){

	//???????????????
	r.POST("./subOrder", func(c *gin.Context) {

		types := c.PostForm("type")
		side := c.PostForm("side")

		//OrderParam
		//????????????
		trader := c.PostForm("trader")
		//fmt.Println(trader)
		//??????
		amount ,_:= strconv.ParseFloat(c.PostForm("amount"),64)
		amount = amount/1000000000000000000
		//fmt.Println(amount)
		//????????????
		data := c.PostForm("data")
		//fmt.Println(data)
		//???????????????
		signature := c.PostForm("signRes")
		//fmt.Println(signature)
		//???????????????
		//sigStruct := c.PostForm("signature")
		//fmt.Println(sigStruct)
		//??????????????????????????????????????????
		perpetual := c.PostForm("perpetual")
		//fmt.Println(perpetual)
		//?????????or?????????
		broker := c.PostForm("broker")
		//fmt.Println(broker)
		//?????????????????????????????????????????????
		orderHash := c.PostForm("orderHash")
		//fmt.Println(orderID)


		if types == "limit"{
			//??????
			price , _:= strconv.ParseFloat(c.PostForm("price"),64)
			price = price/1000000000000000000
			if side == "sell"{
				param ,_ , _ ,_ ,_:=orderBook.ProcessLimitOrder(Sell,orderHash,trader,decimal.NewFromFloat(amount),decimal.NewFromFloat(price),data,signature,perpetual,broker)
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusOK,
					"msg":"ok",
					"data":param,
				})
			}else if side == "buy"{
				param ,_ , _ ,_ ,_:=orderBook.ProcessLimitOrder(Buy,orderHash,trader,decimal.NewFromFloat(amount),decimal.NewFromFloat(price),data,signature,perpetual,broker)
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusOK,
					"msg":"ok",
					"data":param,
				})
			}else{
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusBadRequest,
					"msg":"side must be buy or sell",
				})
			}
		}else if types == "market"{
			price , _:= strconv.ParseFloat(c.PostForm("price"),64)
			price = price/1000000000000000000

			if side == "sell"{
				param ,_ , _ ,_,_,_:= orderBook.ProcessMarketOrder(Sell,orderHash,trader,decimal.NewFromFloat(amount),decimal.NewFromFloat(price),data,signature,perpetual,broker)
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusOK,
					"msg":"ok",
					"data" :param,
				})
			}else if side == "buy"{
				param ,_ , _ ,_,_,_:=orderBook.ProcessMarketOrder(Buy,orderHash,trader,decimal.NewFromFloat(amount),decimal.NewFromFloat(price),data,signature,perpetual,broker)
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusOK,
					"msg":"ok",
					"data":param,
				})
			}else{
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusBadRequest,
					"msg":"side must be buy or sell",
				})
			}
		}else if types == "cancelOrder"{
			if orderBook.IsOwner(orderHash,trader){
				orderBook.CancelOrder(orderHash)
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusOK,
					"msg":"ok",
					"data" : "cancelOrder successful",
				})
			}else{
				c.JSON(http.StatusOK , gin.H{
					"code":http.StatusBadRequest,
					"msg":"error",
					"data" : "you are not owner",
				})
			}
		}else{
			c.JSON(http.StatusOK , gin.H{
				"code":http.StatusBadRequest,
				"msg":"type must be limit , market , cancelOrder",
			})
		}
	})
}