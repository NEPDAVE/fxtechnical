package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"sync"
	"time"
	"strconv"
	"math"
)

/*
-An order is the instruction to buy or sell a currency at a specified rate. The order remains valid until executed or cancelled.
-
-A trade is the execution of the order.
-
-A position is the total of all trades for a specific market.
-*/

/*
***************************
Dragons is a trading algorithm that implements the London daybreak strategy
***************************
*/

//Dragons holds the trading alogorithm state and neccesary data
type Dragons struct {
	Instrument   string
	High         float64 //high from the last three hours
	Low          float64 //low from the last three hours
	LongOrderID  string  //OrderID of current order
	ShortOrderID string
	LongOrders   oanda.ClientOrders //Order SL/TP Limit/Market data
	ShortOrders  oanda.ClientOrders //Order SL/TP Limit/Market data
	SideFilled   SideFilled         //no side/long/short
}

//Init kicks off the goroutines to create orders and check orders
func (d Dragons) Init(instrument string, units string) {
	var wg sync.WaitGroup

	longUnits := units

	//making sure to add -(negative sign) to denote short order
	shortUnits := "-" + units

	//getting and unmarshaling last three hourly candle data
	candles, err := Candles(instrument, "3", "H1")

	if err != nil {
		log.Println(err)
	}

	//getting the previous three hour high and low
	d.High, d.Low = HighAndLow(candles)

	//getting the current bid and ask
	bid, ask := BidAsk("GBP_USD")

  //converting bid to float
	fBid, err := strconv.ParseFloat(bid, 64)

	if err != nil{
		log.Println(err)
	}

	//converting ask to float
	fAsk, err := strconv.ParseFloat(ask, 64)

	if err != nil{
		log.Println(err)
	}

	bidDiff := math.Abs(fBid - d.Low)
	askDiff := math.Abs(fAsk - d.High)

	if fAsk >= d.High {
		fmt.Printf("Ask is higher than previous three hour high by %.5f:\n", askDiff )
    //place MarketLongOrder
	} else if fAsk < d.High{
		fmt.Printf("Ask is lower than previous three hour high by %.5f:\n", askDiff)
    //place LimitLongOrder
	}else{
		fmt.Println("wtf ask")
	}

	if fBid <= d.Low {
		fmt.Printf("Bid is lower than previous three hour low by %.5f:\n", bidDiff)
    //place MarketShortOrder
	} else if fBid > d.Low{
		fmt.Printf("Bid is higher than previous three hour low by %.5f:\n", bidDiff)
    //place LimitShortOrder
	}else{
		fmt.Println("wtf bid")
	}

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

	d.LongOrderID = CreateClientOrdersAndGetOrderID(instrument, longUnits, d.LongOrders)
	d.ShortOrderID = CreateClientOrdersAndGetOrderID(instrument, shortUnits, d.ShortOrders)

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
}

//HandleLongOrder creates either a LimitLongOrder or a MarketLongOrder
//depending on the current Ask in relation to the previous three hour high
func (d *Dragons) HandleLongOrder(orderType string, ask float64, high float64, units string) {

  if orderType == "LIMIT" {


  }

}

//HandleShortOrder creates either a LimitShortOrder or a MarketShortOrder
//depending on the current Bid in relation to the previous three hour low
func (d *Dragons) HandleShortOrder(orderType string, bid float64, low float64, units string) {

}
