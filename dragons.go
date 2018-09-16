package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"math"
)

/*
-An order is the instruction to buy or sell a currency at a specified rate. The order remains valid until executed or cancelled.
-
-A trade is the execution of the order.
-
-A position is the total of all trades for a specific market.

new flow for algorithm

init will set all the needed variables

In the first instance, an order is a request to make a trade to open a position.

A trade is made when the order is matched to a counterparty, ie if you are a buyer, you've found a seller to sell to you, or vice versa.

Once a trade is opened, you hold a position. A position is exposure to the market and will move the balance in your account up or down in line with market movements.

Finally you place an order to close a position which will result in an trade opposite to the direction you initially took, eg if you initially bought, now you sell to close.

And now you hold no position.
*/

/*
***************************
Dragons is a trading algorithm that implements the London daybreak strategy
***************************
*/

//Dragons holds the trading alogorithm state and neccesary data
type Dragons struct {
	Instrument         string
	LongUnits          string
	ShortUnits         string
	High               float64 //high from the last three hours
	Low                float64 //low from the last three hours
	Bid                float64 //current highest Bid
	Ask                float64 //current lowest Ask
	BidDiff            float64 //abv difference between the Bid and the Low
	AskDiff            float64 //abv difference between the Ask and the High
	MarketOrderCreated bool
	LongOrders         OrderData
	ShortOrders        OrderData
}

//OrderData holds the key information about the Order
type OrderData struct {
	Units                  string             //number of units to trade
	OrderID                string             //OrderID of current long order
	TradeID                string             //FIXME TradeID of Order turned Trade?
	Orders                 oanda.ClientOrders //Order SL/TP Limit/Market data
	OrderCreateTransaction *oanda.OrderCreateTransaction
}

//Init kicks off the algorithm to create orders, check orders/trades and cancel
//orders and/or close trades
func (d Dragons) Init(instrument string, units string) {
	d.Instrument = instrument
	d.LongUnits = units
	d.ShortUnits = "-" + units //adding -(negative sign) to denote short order
	d.SetHighAndLow()
	d.SetBidAsk()
	d.BidDiff = math.Abs(d.Bid - d.Low)
	d.AskDiff = math.Abs(d.Ask - d.High)
	d.CreateLongOrders()  //decides to create limit or market orders and returns an OrderCreateTransaction
	d.CreateShortOrders() //decides to create limit or market orders and returns an OrderCreateTransaction

	d.HandleLongOrders()
	d.HandleShortOrders()

	//FIXME need to add trailing stops to order preparations
	//FIXME need to work on order structure and making sure targetPrice/tp/sl
	//data is optimal
	//FIXME need to "handle" orderCreateTransactions vs orderFillTransactions to
	//know whether to check on an order or a trade
	//FIXME call function that tells if it's still an order or a trade...
	//FIXME if needed call function to cancel opposite order yo!
	//FIXME call func to monitor situation. did we win or loose?

	fmt.Printf("%s dragons in the air\n", instrument)
}

//SetHighAndLow sets the previous three hour High and Low for the Dragons struct
func (d *Dragons) SetHighAndLow() {
	//getting and unmarshaling last three hourly candle data
	candles, err := Candles(d.Instrument, "3", "H1")

	if err != nil {
		log.Println(err)
	}

	//getting the previous three hour high and low
	d.High, d.Low = HighAndLow(candles)
}

//SetBidAsk sets the current Bid and Ask for the Dragons struct
//FIXME it may be wise to use the highest ask and lowest bid for this because
//those prices will have the highest liquidity...
func (d *Dragons) SetBidAsk() {
	//getting the current highest bid and  lowest ask
	pricesData := PricesData{}.Init(d.Instrument)
	d.Bid = pricesData.Bid
	d.Ask = pricesData.Ask
}

//CreateLongOrder creates either a LimitLongOrder or a MarketLongOrder
//depending on the current Ask in relation to the previous three hour high
func (d *Dragons) CreateLongOrders() {
	//below if/else if conditional checks current market prices and then places
	//the correct type of Long Order depening on current price action above or
	//below the previous three hour high

	//checking if the current Ask is higher than the  previous three hour high
	//and that no Market Order has already been placed. IE this means the price
	//action is headed "up" and has already broken the three previous three hour
	//high. To ride the trend "up" a Long Market Order is placed.
	if d.Ask >= d.High && d.MarketOrderCreated == false {
		fmt.Printf("Ask is higher than previous three hour high by %.5f:\n", d.AskDiff)

		//preparing the market long order
		d.LongOrders.Orders = MarketLongOrder(d.Bid, d.Ask, d.Instrument, d.LongUnits)

		//creating the order and returning an oanda.OrderCreateTransaction
		d.LongOrders.OrderCreateTransaction = CreateClientOrders(d.Instrument,
			d.LongUnits, d.LongOrders.Orders)

		//making field true to signify a only a long position is being taken IE
		//no need to place a Short Order for the day because the trend is "up"
		d.MarketOrderCreated = true

		//checking if the current Ask is lower than the previous three hour high. IE
		//the price action has not yet broken the previous three hour high so a Limit
		// Long Order is placed in case the trend does go "up"
	} else if d.Ask < d.High {
		fmt.Printf("Ask is lower than previous three hour high by %.5f:\n", d.AskDiff)

		//setting the limit long order targetPrice one pip above the thre hour high
		targetPrice := (d.High + .001) //FIXME need to work on order structure...

		//preparing the Limit Long Order
		d.LongOrders.Orders = LimitLongOrder(targetPrice, d.Instrument, d.LongUnits)

		//creating the Order and returning an oanda.OrderCreateTransaction
		d.LongOrders.OrderCreateTransaction = CreateClientOrders(d.Instrument,
			d.LongUnits, d.LongOrders.Orders)
	}
}

//CreateShortOrder creates either a LimitShortOrder or a MarketShortOrder
//depending on the current Bid in relation to the previous three hour low
func (d *Dragons) CreateShortOrders() {
	//below if/else if conditional checks current market prices and then places
	//the correct type of Short Order depening on current price action above or
	//below the previous three hour low

	//checking if the current Bid is lower than the  previous three hour low
	//and that no Market Order has already been placed. IE this means the price
	//action is headed "down" and has already broken the three previous three hour
	//low. To ride the trend "down" a Short Market Order is placed.
	if d.Bid <= d.Low && d.MarketOrderCreated == false {
		fmt.Printf("Bid is lower than previous three hour low by %.5f:\n", d.BidDiff)

		//preparing the Market Short Order
		d.ShortOrders.Orders = MarketShortOrder(d.Bid, d.Ask, d.Instrument, d.ShortUnits)

		//creating the Order and returning an oanda.OrderCreateTransaction
		d.ShortOrders.OrderCreateTransaction = CreateClientOrders(d.Instrument, d.ShortUnits,
			d.ShortOrders.Orders)

		//making field true to signify a only a short position is being taken IE
		//no need to place a Long Order for the day because the trend is "down"
		d.MarketOrderCreated = true

		//checking if the current Bid is higher than the previous three hour low. IE
		//the price action has not yet broken the previous three hour low so a Limit
		//Short Order is placed in case the trend does go "down"
	} else if d.Bid > d.Low {
		fmt.Printf("Bid is higher than previous three hour low by %.5f:\n", d.BidDiff)

		//setting the limit short order targetPrice one pip below the three hour low
		targetPrice := (d.Low - .001) //FIXME need to work on order structure

		//preparing the Limig Short Order
		d.ShortOrders.Orders = LimitShortOrder(targetPrice, d.Instrument, d.ShortUnits)

		//creating the Order and returning an oanda.OrderCreateTransaction
		d.ShortOrders.OrderCreateTransaction = CreateClientOrders(d.Instrument, d.ShortUnits,
			d.ShortOrders.Orders)
	}
}

//HandleLongOrders uses the data in the d.LongOrders.OrderCreateTransaction to
//determine whether to monitor an order or a trade and to if neccesary cancel
//an oppososite limit order
func (d *Dragons) HandleLongOrders() {
	fmt.Println("Long OrderCreateTransaction:")
	fmt.Println(d.LongOrders.OrderCreateTransaction)
	d.LongOrders.OrderID = d.LongOrders.OrderCreateTransaction.
		OrderCreateTransaction.ID
	state := GetOrderState(d.LongOrders.OrderID)
	fmt.Println(state)
	fmt.Println("")
	d.LongOrders.TradeID = d.LongOrders.OrderCreateTransaction.
		OrderFillTransaction.TradeOpened.TradeID
	fmt.Println("TradeID")
	fmt.Println(d.LongOrders.TradeID)

}

//HandleShortOrders uses the data in the d.ShortOrders.OrderCreateTransaction to
//determine whether to monitor an order or a trade and to if neccesary cancel
//an oppososite limit order
func (d *Dragons) HandleShortOrders() {
	fmt.Println("Short OrderCreateTransaction:")
	fmt.Println(d.ShortOrders.OrderCreateTransaction)
	d.ShortOrders.OrderID = d.ShortOrders.OrderCreateTransaction.
		OrderCreateTransaction.ID
	state := GetOrderState(d.ShortOrders.OrderID)
	fmt.Println(state)
	fmt.Println("")
	d.ShortOrders.TradeID = d.ShortOrders.OrderCreateTransaction.
		OrderFillTransaction.TradeOpened.TradeID
	fmt.Println("TradeID")
	fmt.Println(d.ShortOrders.TradeID)


}

/*


//setting the long limit order target price
	//longTargetPrice := (d.High + .001)

	//setting the long target price far below current ask
	longTargetPrice := (fBid - .010)
	d.LongOrders = LimitLongOrder(longTargetPrice, instrument, units)

	//setting the short limit order target price
	//shortTargetPrice := (d.Low - .001)

	//setting the short target price far above the current bid
	shortTargetPrice := (fAsk + .010)
	d.ShortOrders = LimitShortOrder(shortTargetPrice, instrument, units)

	d.LongOrderID = CreateClientOrdersAndGetOrderID(instrument, d.LongUnits, d.LongOrders)
	d.ShortOrderID = CreateClientOrdersAndGetOrderID(instrument, d.ShortUnits, d.ShortOrders)

	fmt.Printf("Long OrderID: %s\n", d.LongOrderID)
	fmt.Printf("Short OrderID: %s\n", d.ShortOrderID)

	longOrderState := GetOrderState(d.LongOrderID)
	shortOrderState := GetOrderState(d.ShortOrderID)
	existing := GetOrderState("10602")

	fmt.Printf("Long Order State: %s\n", longOrderState)
	fmt.Printf("Short Order State: %s\n", shortOrderState)
	fmt.Printf("Existing Order State: %s\n", existing)

	OrderStateChan := make(chan OrderState)

	wg.Add(2)
	go ContinuousGetOrder(d.LongOrderID, OrderStateChan)
	go ContinuousGetOrder(d.ShortOrderID, OrderStateChan)

	for orderState := range OrderStateChan {
		fmt.Printf("Long OrderID %s State: %s\n", d.LongOrderID, orderState.State)
		fmt.Printf("Short OrderID %s State: %s\n", d.ShortOrderID, orderState.State)
		//
		// 	// 	if orderState.OrderID == d.LongOrderID {
		// 	// 		d.HandleLongOrder(orderState)
		// 	// 	}
		// 	//
		// 	// 	if orderState.State == d.ShortOrderID {
		// 	// 		d.HandleShortOrder(orderState)
		// 	// 	}
	}

	wg.Wait()

*/
