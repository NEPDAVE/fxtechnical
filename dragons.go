package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	twilio "github.com/nepdave/twilio"
	"log"
	"math"
	"time"
	"sync"
	"os"
	//"io/ioutil"
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
with a single market order instead of two limit orders
***************************
*/

//Dragons holds the trading alogorithm state and neccesary data
type Dragons struct {
	wg sync.WaitGroup
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
	TimeOut            bool //program runs for four hours if no trade is placed
	LongOrders         OrderData
	ShortOrders        OrderData
}

//Init kicks off the methods to create orders and check orders
func (d Dragons) Init(instrument string, units string) {
	d.Instrument = instrument
	d.LongUnits = units
	d.ShortUnits = "-" + units //adding -(negative sign) to denote short order
	d.SetHighAndLow()
	d.SetBidAsk()
	d.BidDiff = math.Abs(d.Bid - d.Low)
	d.AskDiff = math.Abs(d.Ask - d.High)
	d.PrepareLongOrders()
	d.PrepareShortOrders()

	d.wg.Add(1)
	go d.TradeTimer()
  d.MonitorPrices()
	d.wg.Wait()
  d.WriteToDoneFile()
	//d.ReadFile()
}

// func (d *Dragons) ReadFile() {
// 	exitingTime, err := ioutil.ReadFile("end.txt")
//
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }

func (d *Dragons) WriteToDoneFile() {
	fmt.Println("Writing to done file...")

	// use touch if log.txt does not exist, 0644 is standard permission
	file, err := os.OpenFile("done.txt", os.O_WRONLY, 0644)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	t := time.Now()
	fmt.Fprintf(file, "Dragons done: %s\n", t.String())
	message := fmt.Sprintf("Dragons done: %s\n", t.String())
	twilio.SendSms("15038411492", message)
}

//TradeTimer set a four hour limit on alogorithm run time. The timer will
//cause the program to end if a placed trade does not.
func (d *Dragons) TradeTimer() {
	timer := time.NewTimer(45 * time.Second)

	<-timer.C
	fmt.Println("Timer 1 expired")
	d.TimeOut = true
	d.wg.Done()
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
	//FIXME should this should be bid and ask with the most liquidity?
	//currently using the highest bid and lowest ask...
	pricesData := PricesData{}.Init(d.Instrument)
	d.Bid = pricesData.Bid
	d.Ask = pricesData.Ask
}

func (d *Dragons) PrepareLongOrders() {
	//setting stop loss 5 pips below the d.Low
	stopLossPrice := fmt.Sprintf("%.5f", (d.Low - .0005))

	//setting take profit 15 pips above the d.High
	takeProfitPrice := fmt.Sprintf("%.5f", (d.High + .0015))

	//building struct needed for marshaling data into a []byte
	d.LongOrders.Orders = MarketOrder(stopLossPrice, takeProfitPrice,
		d.Instrument, d.LongUnits)

	//marshaling the struct into a byte slice for order creation
	d.LongOrders.OrdersByte = oanda.ClientOrders{}.MarshalClientOrders(
		d.LongOrders.Orders)

	fmt.Println("Long Orders:")
	fmt.Println(string(d.LongOrders.OrdersByte))
}

func (d *Dragons) PrepareShortOrders() {
	//setting stop loss 5 pips above the d.High
	stopLossPrice := fmt.Sprintf("%.5f", (d.Low + .0005))

	//setting take profit 15 pips below the d.Low
	takeProfitPrice := fmt.Sprintf("%.5f", (d.High - .0015))

	//building struct needed for marshaling data into a []byte
	d.ShortOrders.Orders = MarketOrder(stopLossPrice, takeProfitPrice,
		d.Instrument, d.ShortUnits)

	//marshaling the struct into a byte slice for order creation
	d.ShortOrders.OrdersByte = oanda.ClientOrders{}.MarshalClientOrders(
		d.ShortOrders.Orders)

	fmt.Println("Short Orders:")
	fmt.Println(string(d.ShortOrders.OrdersByte))
}

func (d *Dragons) MonitorPrices() {
	//if a market order has not been created loop continues and the timer has
	//not run out the loop continues
	fmt.Println("Entering MonitorPrices loop...")
	for d.MarketOrderCreated == false && d.TimeOut == false {
		d.SetBidAsk()
		// fmt.Println("#######################")
		//fmt.Println(time.Now())
		// fmt.Printf("Highest Bid: %f\n", d.Bid)
		// fmt.Printf("BidDiff ABV: %.5f\n", d.BidDiff)
		// fmt.Println("")
		// fmt.Printf("Lowest Ask: %f\n", d.Ask)
		// fmt.Printf("AskDiff ABV: %.5f\n", d.AskDiff)
		// fmt.Println("")
		// fmt.Printf("Spread: %.5f\n", (d.Ask - d.Bid))

		if d.Ask > d.High {
			fmt.Println("Going long!")
			createOrdersByte, err := oanda.CreateOrder(d.LongOrders.OrdersByte)

			if err != nil {
				log.Println(err)
			}

			fmt.Println(string(createOrdersByte))
			d.MarketOrderCreated = true
			return

		} else if d.Bid < d.Low {
			fmt.Println("Going short!")
			createOrdersByte, err := oanda.CreateOrder(d.ShortOrders.OrdersByte)

			if err != nil {
				log.Println(err)
			}

			fmt.Println(string(createOrdersByte))
			d.MarketOrderCreated = true
			return
   }
	}
}
