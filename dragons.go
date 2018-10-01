package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	twilio "github.com/nepdave/twilio"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
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
with a single market order
***************************
*/

//Dragons holds the trading alogorithm state and neccesary data
type Dragons struct {
	Instrument             string
	LongUnits              string
	ShortUnits             string
	High                   float64 //high from the last three hours
	Low                    float64 //low from the last three hours
	Bid                    float64 //current highest Bid
	Ask                    float64 //current lowest Ask
	BidDiff                float64 //abv difference between the Bid and the Low
	AskDiff                float64 //abv difference between the Ask and the High
	MarketOrderCreated     bool
	TimeOut                bool //program runs for four hours if no trade is placed
	LongOrders             OrderData
	ShortOrders            OrderData
	OrderCreateTransaction string
}

//Init kicks off the methods to create orders and check orders
func (d Dragons) Init(instrument string, units string) {
	d.Instrument = instrument
	d.LongUnits = units
	d.ShortUnits = "-" + units //adding -(negative sign) to denote short order
	d.SetHighAndLow()
	d.BidDiff = math.Abs(d.Bid - d.Low)
	d.AskDiff = math.Abs(d.Ask - d.High)
	d.MonitorPrices()
	d.WriteToDoneFile()
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

	takeProfitPriceFloat := 3 * ((d.High - d.Low) + .0005)
	takeProfitPrice := fmt.Sprintf("%.5f", (d.Ask + takeProfitPriceFloat))

	//building struct needed for marshaling data into a []byte
	d.LongOrders.Orders = MarketOrder(stopLossPrice, takeProfitPrice,
		d.Instrument, d.LongUnits)

	//marshaling the struct into a byte slice for order creation
	d.LongOrders.OrdersByte = oanda.ClientOrders{}.MarshalClientOrders(
		d.LongOrders.Orders)

	fmt.Println("Long Orders:")
	fmt.Println(string(d.LongOrders.OrdersByte))
	fmt.Println("")
}

func (d *Dragons) PrepareShortOrders() {
	//setting stop loss 5 pips above the d.High
	stopLossPrice := fmt.Sprintf("%.5f", (d.High + .0005))

	takeProfitPriceFloat := 3 * ((d.High - d.Low) + .0005)
	takeProfitPrice := fmt.Sprintf("%.5f", (d.Bid - takeProfitPriceFloat))

	//building struct needed for marshaling data into a []byte
	d.ShortOrders.Orders = MarketOrder(stopLossPrice, takeProfitPrice,
		d.Instrument, d.ShortUnits)

	//marshaling the struct into a byte slice for order creation
	d.ShortOrders.OrdersByte = oanda.ClientOrders{}.MarshalClientOrders(
		d.ShortOrders.Orders)

	fmt.Println("Short Orders:")
	fmt.Println(string(d.ShortOrders.OrdersByte))
	fmt.Println("")
}

//MonitorPrices checks that the timer has not run out and that an order has not
//been created and continues to MonitorPrices for a breakout
func (d *Dragons) MonitorPrices() {
	var wg sync.WaitGroup
	var timer = time.NewTimer(45 * time.Second)

	wg.Add(1)
	go func() {
		//when the Timer expires, the current time will be sent on C indicating
		//the Timer is done
		<-timer.C
		d.TimeOut = true
		wg.Done()
	}()

	//if a market order has not been created loop continues and the timer has
	//not run out the loop continues
	fmt.Println("Entering MonitorPrices loop...")
	fmt.Println("")
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
			d.PrepareLongOrders()
			createOrdersByte, err := oanda.CreateOrder(d.LongOrders.OrdersByte)

			if err != nil {
				log.Println(err)
			}

			d.OrderCreateTransaction = string(createOrdersByte)
			fmt.Println("Long Order Create Transaction:")
			fmt.Println(d.OrderCreateTransaction)
			fmt.Println("")

			d.MarketOrderCreated = true

			timer.Stop()
			wg.Done()
			return

		} else if d.Bid < d.Low {
			d.PrepareShortOrders()
			createOrdersByte, err := oanda.CreateOrder(d.ShortOrders.OrdersByte)

			if err != nil {
				log.Println(err)
			}

			d.OrderCreateTransaction = string(createOrdersByte)
			fmt.Println("Short Order Create Transaction:")
			fmt.Println(d.OrderCreateTransaction)
			fmt.Println("")

			d.MarketOrderCreated = true

			timer.Stop()
			wg.Done()
			return
		}
	}
	wg.Wait()
}

func (d *Dragons) WriteToDoneFile() {
	fmt.Println("Writing to done.txt...")

	// use touch if log.txt does not exist, 0644 is standard permission
	file, err := os.OpenFile("done.txt", os.O_WRONLY, 0644)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	fmt.Fprintf(file, "Dragons done: %s\n", time.Now().String())
	done := fmt.Sprintf("Dragons done: %s\n", time.Now().String())

	marketOrderCreated := fmt.Sprintf("Market Order Created: %s\n",
		strconv.FormatBool(d.MarketOrderCreated))

	orderCreateTransaction := fmt.Sprintf("Order Create Transaction: %s\n",
		d.OrderCreateTransaction)

	message := done + marketOrderCreated + orderCreateTransaction

	twilio.SendSms("15038411492", message)
}
