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

new flow for alogorithm
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
	Orders           oanda.ClientOrders //Order SL/TP Limit/Market data
	OrderCreateTransaction oanda.OrderCreateTransaction
}

//Init kicks off the goroutines to create orders and check orders
func (d Dragons) Init(instrument string, units string) {
	d.Instrument = instrument
	d.LongUnits = units
	d.ShortUnits = "-" + units //adding -(negative sign) to denote short order
	d.SetHighAndLow()
	d.SetBidAsk()
	d.BidDiff = math.Abs(d.Bid - d.Low)
	d.AskDiff = math.Abs(d.Ask - d.High)
	d.CreateLongOrders() //decides to create limit or market orders and returns an OrderCreateTransaction
	d.CreateShortOrders() //decides to create limit or market orders and returns an OrderCreateTransaction

	//d.HandleLongOrders(orderCreateTransaction) //uses the data in the
	//OrderCreateTransaction to determine whether to monitor an order or a trade

	//d.HandleShortOrders(orderCreateTransaction) //uses the data in the
	//OrderCreateTransaction to determine whether to monitor an order or a trade



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
func (d *Dragons) SetBidAsk() {
	//getting the current highest bid and  lowest ask
	pricesData := PricesData{}.Init(d.Instrument)
	d.Bid = pricesData.Bid
	d.Ask = pricesData.Ask
}

//HandleLongOrder creates either a LimitLongOrder or a MarketLongOrder
//depending on the current Ask in relation to the previous three hour high
func (d *Dragons) CreateLongOrders() {
	//checking current market prices and then placing the correct type of Long
	//depening on current price action
	if d.Ask >= d.High && d.MarketOrderCreated == false {
		fmt.Printf("Ask is higher than previous three hour high by %.5f:\n", d.AskDiff)

		//place MarketLongOrder
		d.LongOrders.Orders = MarketLongOrder(d.Bid, d.Ask, d.Instrument, d.LongUnits)
		d.LongOrders.OrderID = CreateClientOrdersAndGetOrderID(d.Instrument,
			d.LongUnits, d.LongOrders.Orders)
		d.MarketOrderCreated = true

	} else if d.Ask < d.High {
		fmt.Printf("Ask is lower than previous three hour high by %.5f:\n", d.AskDiff)

		//place LimitLongOrder
		targetPrice := (d.High + .001) //FIXME need to work on order structure...
		d.LongOrders.Orders = LimitLongOrder(targetPrice, d.Instrument, d.LongUnits)
		d.LongOrders.OrderID = CreateClientOrdersAndGetOrderID(d.Instrument,
			d.LongUnits, d.LongOrders.Orders)

	} else {
		fmt.Println("wtf ask...")
	}

}

//HandleShortOrder creates either a LimitShortOrder or a MarketShortOrder
//depending on the current Bid in relation to the previous three hour low
func (d *Dragons) CreateShortOrders() {
	//checking current market prices and then placing the correct type of Short
	//depening on current price action
	if d.Bid <= d.Low && d.MarketOrderCreated == false {
		fmt.Printf("Bid is lower than previous three hour low by %.5f:\n", d.BidDiff)

		//place MarketShortOrder
		d.ShortOrders.Orders = MarketShortOrder(d.Bid, d.Ask, d.Instrument, d.ShortUnits)
		d.ShortOrders.OrderID = CreateClientOrdersAndGetOrderID(d.Instrument,
			d.ShortUnits, d.ShortOrders.Orders)
		d.MarketOrderCreated = true

	} else if d.Bid > d.Low {
		fmt.Printf("Bid is higher than previous three hour low by %.5f:\n", d.BidDiff)

		//place LimitShortOrder
		targetPrice := (d.Low - .001) //FIXME need to work on order structure
		d.ShortOrders.Orders = LimitShortOrder(targetPrice, d.Instrument, d.ShortUnits)
		d.ShortOrders.OrderID = CreateClientOrdersAndGetOrderID(d.Instrument,
			d.ShortUnits, d.ShortOrders.Orders)

	} else {
		fmt.Println("wtf bid")
	}

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
