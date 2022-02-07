package main

import (
	"container/list"
	"encoding/json"
	"github.com/shopspring/decimal"
	"sort"
	"strconv"
	"time"
)

type OrderBook struct {
	orders map[string]*list.Element // orderID -> *Order (*list.Element.Value.(*Order))
	asks *OrderSide
	bids *OrderSide
    historyDeal []*HistoryDeal
	traderDeal map[string]*TraderPosition
	handicap *Handicap
}

type Handicap struct {
	BidsHandicap []*OrderQueue `json:"bids"`
	AsksHandicap []*OrderQueue `json:"asks"`
}
func (ob *OrderBook) setHandicap()  {
	var bidsHandicap []*OrderQueue
	var asksHandicap []*OrderQueue
	var bidSlice []float64
	var askSlice []float64
	for k,_ := range ob.bids.prices{
		key,_ := strconv.ParseFloat(k,64)
		bidSlice = append(bidSlice,key)
	}
	sort.Float64s(bidSlice)
	for k,_ := range ob.asks.prices{
		key,_ := strconv.ParseFloat(k,64)
		askSlice = append(askSlice,key)
	}
	sort.Float64s(askSlice)

	if bidSlice != nil{
		for i := len(ob.bids.prices)-1 ;i>=0 ; i--{
			bidsHandicap = append(bidsHandicap, ob.bids.prices[strconv.FormatFloat(bidSlice[i],'f',-1,64)])
		}
	}
	if askSlice != nil{
		for i := len(ob.asks.prices)-1 ;i>=0 ; i-- {
			asksHandicap = append(asksHandicap, ob.asks.prices[strconv.FormatFloat(askSlice[i],'f',-1,64)])
		}
	}
	ob.handicap.BidsHandicap = bidsHandicap
	ob.handicap.AsksHandicap = asksHandicap
}

type TraderPosition struct {
	Ask *Order `json:"ask"`
	Bid *Order `json:"bid"`
}

func NewPosition() *TraderPosition {
	return &TraderPosition{
		Ask: &Order{},
		Bid: &Order{},
	}
}

type HistoryDeal struct {
	Side   Side            `json:"side"`
	Price  decimal.Decimal `json:"price"`
	Amount decimal.Decimal `json:"amount"`
	Time   string       `json:"time"`
}

type Param struct{
	TakerParam *Order   `json:"taker_Param"`
	MakerParam []*Order `json:"maker_Param"`
}


func NewOrderParam() *Param{
	return &Param{
		TakerParam: &Order{},
		MakerParam: []*Order{},
	}
}

// NewHistoryDeal 对于限价单来说，挂单吃单都在done里面,done的类型[]*Order
//历史成交需要数量，价格，时间，从done里取出数量及价格，存入数据结构，加上时间戳
// 我的订单，时间，类型，买/卖，数量，价格，合计，挂单/吃单
func NewHistoryDeal(side Side,price , amount decimal.Decimal,time string) *HistoryDeal{
	return &HistoryDeal{
		Side:   side,
		Price:  price,
		Amount: amount,
		Time:   time,
	}
}

// NewOrderBook creates Orderbook object
func NewOrderBook() *OrderBook {
	return &OrderBook{
		orders: map[string]*list.Element{},
		bids:   NewOrderSide(),
		asks:   NewOrderSide(),
		//traderDeal: map[string][]*Order{},
		traderDeal: map[string]*TraderPosition{},
		handicap: &Handicap{
			BidsHandicap: []*OrderQueue{},
			AsksHandicap: []*OrderQueue{},
		},
	}
}


// PriceLevel contains price and volume in depth
type PriceLevel struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
}

// ProcessMarketOrder 下市价单
func (ob *OrderBook) ProcessMarketOrder(side Side,orderID , trader string , amount ,price decimal.Decimal,data,signature,perpetual,broker string) (orderParam *Param , done []*Order, partial *Order, partialQuantityProcessed, quantityLeft decimal.Decimal, err error) {
	quantity := amount
	if quantity.Sign() <= 0 {
		return nil,nil, nil, decimal.Zero, decimal.Zero, ErrInvalidQuantity
	}

	var (
		iter          func() *OrderQueue
		sideToProcess *OrderSide
	)

	if side == Buy {
		iter = ob.asks.MinPriceQueue
		sideToProcess = ob.asks
	} else {
		iter = ob.bids.MaxPriceQueue
		sideToProcess = ob.bids
	}
	makerQuantityAll := decimal.Zero
	quantityMulAmount := decimal.Zero
	orderParam = NewOrderParam()
	year := strconv.Itoa(time.Now().Year())
	month := strconv.Itoa(int(time.Now().Month()))
	if len(month[:]) == 1{
		month = "0"+month
	}
	day := strconv.Itoa(time.Now().Day())
	if len(day[:]) == 1{
		day = "0"+day
	}
	hour := strconv.Itoa(time.Now().Hour())
	if len(hour[:]) == 1{
		hour = "0"+hour
	}
	min := strconv.Itoa(time.Now().Minute())
	if len(min[:]) == 1{
		min = "0"+min
	}
	sec := strconv.Itoa(time.Now().Second())
	if len(sec[:]) == 1{
		sec = "0"+sec
	}
	timeNow := year + "-" + month + "-" + day + " " + hour + ":" + min +":"+sec

	for quantity.Sign() > 0 && sideToProcess.Len() > 0 {
		//迭代最优价格，bestPrice是一个价格队列
		bestPrice := iter()
		//相同价格买卖的数量
		//samePriceQuantity := decimal.Zero
		//获取买卖结果
		ordersDone, partialDone, partialProcessed, quantityLeft := ob.processQueue(bestPrice, quantity)
		//添加结果到done和orderParam.MakerParam，这个ordersDone是最优价格的全部完成订单
		orderParam.MakerParam = append(orderParam.MakerParam, ordersDone...)
		done = append(done, ordersDone...)
		//该价格队列没有交易者部分匹配的订单
		if partialDone == nil{
			makerQuantityAll = makerQuantityAll.Add(quantity.Sub(quantityLeft))
			quantityMulAmount = quantityMulAmount.Add((quantity.Sub(quantityLeft)).Mul(ordersDone[0].price))

		}else {
			makerQuantityAll = makerQuantityAll.Add(quantity.Sub(quantityLeft))
			quantityMulAmount = quantityMulAmount.Add((quantity.Sub(quantityLeft)).Mul(partialDone.price))
			doneAmountAll := decimal.Zero
			for _,v := range ordersDone{
				doneAmountAll = doneAmountAll.Add(v.quantity)
			}
			orderParam.MakerParam = append(orderParam.MakerParam , partialDone)
		}
		partial = partialDone
		partialQuantityProcessed = partialProcessed
		quantity = quantityLeft
	}
	if !makerQuantityAll.IsZero(){
		orderParam.TakerParam = NewOrder(orderID,side,trader,amount,price,timeNow,data,signature,perpetual,broker)

		for _ ,v := range orderParam.MakerParam{
			ob.historyDeal = append(ob.historyDeal, NewHistoryDeal(v.side,v.price,v.quantity,timeNow))
		}
	}else{
		orderParam = nil
	}
	ob.setHandicap()
	quantityLeft = quantity
	return
}

// ProcessLimitOrder 下限价单
func (ob *OrderBook) ProcessLimitOrder( side Side, orderID string,trader string, quantity, price decimal.Decimal,data,signature,perpetual,broker string) (orderParam *Param,done []*Order, partial *Order, partialQuantityProcessed decimal.Decimal, err error) {
	if _, ok := ob.orders[orderID]; ok {
		return nil, nil, nil, decimal.Zero, ErrOrderExists
	}

	if quantity.Sign() <= 0 {
		return nil,nil, nil, decimal.Zero, ErrInvalidQuantity
	}

	if price.Sign() <= 0 {
		return nil,nil, nil, decimal.Zero, ErrInvalidPrice
	}

	quantityToTrade := quantity
	makerQuantityAll := decimal.Zero
	quantityMulAmount := decimal.Zero
	var (
		sideToProcess *OrderSide
		sideToAdd     *OrderSide
		comparator    func(decimal.Decimal) bool
		iter          func() *OrderQueue
	)

	if side == Buy {
		sideToAdd = ob.bids
		sideToProcess = ob.asks
		comparator = price.GreaterThanOrEqual
		iter = ob.asks.MinPriceQueue
	} else {
		sideToAdd = ob.asks
		sideToProcess = ob.bids
		comparator = price.LessThanOrEqual
		iter = ob.bids.MaxPriceQueue
	}


	bestPrice := iter()
	orderParam = NewOrderParam()
	year := strconv.Itoa(time.Now().Year())
	month := strconv.Itoa(int(time.Now().Month()))
	if len(month[:]) == 1{
		month = "0"+month
	}
	day := strconv.Itoa(time.Now().Day())
	if len(day[:]) == 1{
		day = "0"+day
	}
	hour := strconv.Itoa(time.Now().Hour())
	if len(hour[:]) == 1{
		hour = "0"+hour
	}
	min := strconv.Itoa(time.Now().Minute())
	if len(min[:]) == 1{
		min = "0"+min
	}
	sec := strconv.Itoa(time.Now().Second())
	if len(sec[:]) == 1{
		sec = "0"+sec
	}
	timeNow := year + "-" + month + "-" + day + " " + hour + ":" + min +":"+sec
	for quantityToTrade.Sign() > 0 && sideToProcess.Len() > 0 && comparator(bestPrice.Price()) {
		ordersDone, partialDone, partialQty, quantityLeft := ob.processQueueForLimit(bestPrice, quantityToTrade)
		//被完整交易的订单
		done = append(done, ordersDone...)
		orderParam.MakerParam = append(orderParam.MakerParam, ordersDone...)

		//部分完成订单不为空，交易要完成了
		if partialDone == nil {
			makerQuantityAll = makerQuantityAll.Add(quantityToTrade.Sub(quantityLeft))
			quantityMulAmount = quantityMulAmount.Add((quantityToTrade.Sub(quantityLeft)).Mul(done[0].price))
		}else {
			//对方部分订单完成,证明全部amount全部卖出/买入
			if partialDone.side != side{
				makerQuantityAll = makerQuantityAll.Add(quantityToTrade)
				quantityMulAmount = quantityMulAmount.Add(quantityToTrade.Mul(partialDone.price))
				doneAmountAll := decimal.Zero
				for _,v := range ordersDone{
					doneAmountAll = doneAmountAll.Add(v.quantity)
				}
				orderParam.MakerParam = append(orderParam.MakerParam, NewOrder(partialDone.id,partialDone.side,partialDone.trader,quantityToTrade.Sub(doneAmountAll),partialDone.price,timeNow,partialDone.data,partialDone.signature,partialDone.perpetual,partialDone.broker))
			}else{
				//交易者未购买完全,就是剩余交易量减去下单量
				makerQuantityAll = makerQuantityAll.Add(quantityToTrade.Sub(partialDone.quantity))
				quantityMulAmount = quantityMulAmount.Add(partialDone.price.Mul(quantityToTrade.Sub(partialDone.quantity)))
			}
		}
		partial = partialDone
		partialQuantityProcessed = partialQty
		quantityToTrade = quantityLeft
		bestPrice = iter()
	}
	//所有单都吃了也没满足交易数量
	if quantityToTrade.Sign() > 0 {
		//用户要将多余的单挂上去
		o := NewOrder(orderID, side, trader,quantityToTrade, price, timeNow,data,signature,perpetual,broker)
		if len(done) > 0 {
			partialQuantityProcessed = quantity.Sub(quantityToTrade)
			partial = o
		}
		//这里的o是剩余订单
		ob.orders[orderID] = sideToAdd.Append(o)
	}
	if !makerQuantityAll.IsZero(){
		orderParam.TakerParam = NewOrder(orderID,side,trader,makerQuantityAll,quantityMulAmount.Div(makerQuantityAll),timeNow,data,signature,perpetual,broker)
		if ob.traderDeal[trader] == nil{
			ob.traderDeal[trader] = NewPosition()
		}
			if side == Sell{
				ob.traderDeal[trader].Ask = NewOrder(orderID,side,trader,ob.traderDeal[trader].Ask.quantity.Add(orderParam.TakerParam.quantity),(ob.traderDeal[trader].Ask.price.Mul(ob.traderDeal[trader].Ask.quantity)).Add(orderParam.TakerParam.quantity.Mul(orderParam.TakerParam.price)).Div(ob.traderDeal[trader].Ask.quantity.Add(orderParam.TakerParam.quantity)),timeNow,data,signature,perpetual,broker)
			}else{
				ob.traderDeal[trader].Bid = NewOrder(orderID,side,trader,ob.traderDeal[trader].Bid.quantity.Add(orderParam.TakerParam.quantity),(ob.traderDeal[trader].Bid.price.Mul(ob.traderDeal[trader].Bid.quantity)).Add(orderParam.TakerParam.quantity.Mul(orderParam.TakerParam.price)).Div(ob.traderDeal[trader].Bid.quantity.Add(orderParam.TakerParam.quantity)),timeNow,data,signature,perpetual,broker)
			}


		for _ ,v := range orderParam.MakerParam{
			ob.historyDeal = append(ob.historyDeal, NewHistoryDeal(v.side,v.price,v.quantity,timeNow))
			if ob.traderDeal[v.trader] == nil{
				ob.traderDeal[v.trader] = NewPosition()
			}
			if v.side == Sell{
				ob.traderDeal[v.trader].Ask = NewOrder(v.id,v.side,v.trader,ob.traderDeal[v.trader].Ask.quantity.Add(v.quantity),(ob.traderDeal[v.trader].Ask.price.Mul(ob.traderDeal[v.trader].Ask.quantity)).Add(v.quantity.Mul(v.price)).Div(ob.traderDeal[v.trader].Ask.quantity.Add(v.quantity)),timeNow,v.data,v.signature,v.perpetual,v.broker)
			}else{
				ob.traderDeal[v.trader].Bid = NewOrder(v.id,v.side,v.trader,ob.traderDeal[v.trader].Bid.quantity.Add(v.quantity),(ob.traderDeal[v.trader].Bid.price.Mul(ob.traderDeal[v.trader].Bid.quantity)).Add(v.quantity.Mul(v.price)).Div(ob.traderDeal[v.trader].Bid.quantity.Add(v.quantity)),timeNow,v.data,v.signature,v.perpetual,v.broker)
			}

		}
	}else{
		orderParam = nil
	}
	ob.setHandicap()
	return
}


func (ob *OrderBook) processQueue(orderQueue *OrderQueue, quantityToTrade decimal.Decimal) (done []*Order, partial *Order, partialQuantityProcessed, quantityLeft decimal.Decimal) {
	quantityLeft = quantityToTrade

	for orderQueue.Len() > 0 && quantityLeft.Sign() > 0 {
		headOrderEl := orderQueue.Head()
		headOrder := headOrderEl.Value.(*Order)

		if quantityLeft.LessThan(headOrder.Quantity()) {
			partial = NewOrder(headOrder.ID(), headOrder.Side(),headOrder.Trader(), headOrder.Quantity().Sub(quantityLeft), headOrder.Price(), headOrder.Time(),headOrder.data,headOrder.signature,headOrder.perpetual,headOrder.broker)
			partialQuantityProcessed = quantityLeft
			orderQueue.Update(headOrderEl, partial)
			quantityLeft = decimal.Zero
		} else {
			quantityLeft = quantityLeft.Sub(headOrder.Quantity())
			done = append(done, ob.CancelOrder(headOrder.ID()))
		}
	}
	return
}

func (ob *OrderBook) processQueueForLimit(orderQueue *OrderQueue, quantityToTrade decimal.Decimal) (done []*Order, partial *Order, partialQuantityProcessed, quantityLeft decimal.Decimal) {
	quantityLeft = quantityToTrade

	for orderQueue.Len() > 0 && quantityLeft.Sign() > 0 {
		headOrderEl := orderQueue.Head()
		headOrder := headOrderEl.Value.(*Order)

		if quantityLeft.LessThan(headOrder.Quantity()) {
			partial = NewOrder(headOrder.ID(), headOrder.Side(),headOrder.Trader(), headOrder.Quantity().Sub(quantityLeft), headOrder.Price(), headOrder.Time(),headOrder.data,headOrder.signature,headOrder.perpetual,headOrder.broker)
			partialQuantityProcessed = quantityLeft
			orderQueue.Update(headOrderEl, partial)
			quantityLeft = decimal.Zero
		} else {
			quantityLeft = quantityLeft.Sub(headOrder.Quantity())
			done = append(done, ob.CancelOrder(headOrder.ID()))
		}
	}
	return
}

// OrderByHash Order returns order by id
func (ob *OrderBook) OrderByHash(orderID string) *Order {
	e, ok := ob.orders[orderID]
	if !ok {
		return nil
	}

	return e.Value.(*Order)
}

func (ob *OrderBook) OrderByAddress(orderAddress string) []*Order {
	var traderOrders []*Order
	for _,v :=range ob.orders {
		if v.Value.(*Order).trader == orderAddress{
			traderOrders = append(traderOrders, v.Value.(*Order))
		}
	}
	return traderOrders
}

// Depth returns price levels and volume at price level
func (ob *OrderBook) Depth() (asks, bids []*PriceLevel) {
	level := ob.asks.MaxPriceQueue()
	for level != nil {
		asks = append(asks, &PriceLevel{
			Price:    level.Price(),
			Quantity: level.Volume(),
		})
		level = ob.asks.LessThan(level.Price())
	}

	level = ob.bids.MaxPriceQueue()
	for level != nil {
		bids = append(bids, &PriceLevel{
			Price:    level.Price(),
			Quantity: level.Volume(),
		})
		level = ob.bids.LessThan(level.Price())
	}
	return
}

// CancelOrder removes order with given ID from the order book
func (ob *OrderBook) CancelOrder(orderID string) *Order {
	e, ok := ob.orders[orderID]
	if !ok {
		return nil
	}

	delete(ob.orders, orderID)

	if e.Value.(*Order).Side() == Buy {
		return ob.bids.Remove(e)
	}

	return ob.asks.Remove(e)
}

// CalculateMarketPrice returns total market price for requested quantity
// if err is not nil price returns total price of all levels in side
func (ob *OrderBook) CalculateMarketPrice(side Side, quantity decimal.Decimal) (price decimal.Decimal, err error) {
	price = decimal.Zero

	var (
		level *OrderQueue
		iter  func(decimal.Decimal) *OrderQueue
	)

	if side == Buy {
		level = ob.asks.MinPriceQueue()
		iter = ob.asks.GreaterThan
	} else {
		level = ob.bids.MaxPriceQueue()
		iter = ob.bids.LessThan
	}

	for quantity.Sign() > 0 && level != nil {
		levelVolume := level.Volume()
		levelPrice := level.Price()
		if quantity.GreaterThanOrEqual(levelVolume) {
			price = price.Add(levelPrice.Mul(levelVolume))
			quantity = quantity.Sub(levelVolume)
			level = iter(levelPrice)
		} else {
			price = price.Add(levelPrice.Mul(quantity))
			quantity = decimal.Zero
		}
	}

	if quantity.Sign() > 0 {
		err = ErrInsufficientQuantity
	}

	return
}

// String implements fmt.Stringer interface
func (ob *OrderBook) String() string {
	return ob.asks.String() + "\r\n------------------------------------" + ob.bids.String()
}

// MarshalJSON implements json.Marshaler interface
func (ob *OrderBook) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		&struct {
			Asks *OrderSide `json:"asks"`
			Bids *OrderSide `json:"bids"`
			//TraderDeal map[string][]*Order `json:"trader_deal"`
		}{
			Asks: ob.asks,
			Bids: ob.bids,
			//TraderDeal : ob.traderDeal,
		},
	)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ob *OrderBook) UnmarshalJSON(data []byte) error {
	obj := struct {
		Asks *OrderSide `json:"asks"`
		Bids *OrderSide `json:"bids"`
		//TraderDeal map[string][]*Order `json:"trader_deal"`
	}{}

	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	ob.asks = obj.Asks
	ob.bids = obj.Bids
	//ob.traderDeal = obj.TraderDeal
	//ob.historyDeal = obj.HistoryDeals
	ob.orders = map[string]*list.Element{}

	for _, order := range ob.asks.Orders() {
		ob.orders[order.Value.(*Order).ID()] = order
	}

	for _, order := range ob.bids.Orders() {
		ob.orders[order.Value.(*Order).ID()] = order
	}

	return nil
}


