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

	//下限价卖单
	r.POST("./subOrder", func(c *gin.Context) {

		types := c.PostForm("type")
		side := c.PostForm("side")

		//OrderParam
		//用户地址
		trader := c.PostForm("trader")
		//fmt.Println(trader)
		//数量
		amount ,_:= strconv.ParseFloat(c.PostForm("amount"),64)
		amount = amount/1000000000000000000
		//fmt.Println(amount)
		//签名数据
		data := c.PostForm("data")
		//fmt.Println(data)
		//签名字符串
		signature := c.PostForm("signRes")
		//fmt.Println(signature)
		//签名结构体
		//sigStruct := c.PostForm("signature")
		//fmt.Println(sigStruct)
		//合约地址，用于判断上哪个合约
		perpetual := c.PostForm("perpetual")
		//fmt.Println(perpetual)
		//上链人or代理人
		broker := c.PostForm("broker")
		//fmt.Println(broker)
		//订单哈希，用于判断是否重复订单
		orderHash := c.PostForm("orderHash")
		//fmt.Println(orderID)


		if types == "limit"{
			//价格
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
			orderBook.CancelOrder(orderHash)
			c.JSON(http.StatusOK , gin.H{
				"code":http.StatusOK,
				"msg":"ok",
				"data" : "cancelOrder successful",
			})
		}else{
			c.JSON(http.StatusOK , gin.H{
				"code":http.StatusBadRequest,
				"msg":"type must be sellMarket , buyMarket , sellLimit , buyLimit , cancelOrder",
			})
		}
	})
}