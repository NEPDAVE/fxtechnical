package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	// twilio "github.com/nepdave/twilio"
	"log"
	"sync"
	"time"
)

//trying to figure out select stuff
//https://play.golang.org/p/HrPk4uEO2tS

/*
***************************
Raider is a trading algorithm that implements the Bolinger Band indicator
***************************
*/

//Raider holds the algo state and current OrderID
type Raider struct {
	mu sync.Mutex
	RaiderStatus int //0 = orders closed. 1 = order submitted. 2 = orders open.
	OrderID      string
}

//Init kicks off the select pattern to check on raider status
func (r Raider) Init(instrument string, units string) {

  //Takes
	RaiderReconChan := make(chan RaiderRecon)
	var wg sync.WaitGroup

	wg.Add(3)
	go r.ExecuteBB(instrument, units, RaiderReconChan)
	go r.CheckConditions("instrument string", "units string", OrdersStatusChan)
	go r.CheckOrder()
  /*
  the three goroutines should still send a value of the channel
	this will make each of them more modular. you can then init
	use the value sent over the channel to and pass it to CheckConditions
	which is a func that checks the current RaiderStatus and decides
	whether or not to submit the order and get the new order ID.
	*/

	for {
		select {
		case raiderRecon = <-RaiderRecon:
			fmt.Println("Received create order signal...")
			if RaiderStatus == 0 {
				fmt.Println("Verifying ...")
					CheckConditions()
				}
		default:
			fmt.Println("no data...")
		}
	}

	wg.Wait()
}

//CheckConditions updates RaiderStatus and submits/creates the orders
func (r *Raider) CheckConditions(instrument string, units string) {
	fmt.Println("Checking Conditions...")
	//checks bollinger band execute signal


	for OrderStatus := range OrdersStatusChan {
		if OrderStatus == 1 {
			r.RaiderStatus = 1
		}
	}
}

//ExecuteBB executes the Raider trading algorithm
func (r *Raider) ExecuteBB(instrument string, units string, RaiderReconChan chan RaiderRecon) {
	bb := BollingerBand{}.Init(instrument, "20", "D")
	var wg sync.WaitGroup

	//anonymous go func executing concurrently to update bb everyday at midnight
	wg.Add(1)
	go func() {
		for {
			now := time.Now()
			if now.Hour() == 00 && now.Minute() == 0 && now.Second() < 5 {
				bb = BollingerBand{}.Init(instrument, "20", "D")
			}
		}
	}()


	//FIXME where do we really want to set the number of units?
	wg.Add(1)
	go RaiderRecon{}.ContinuousRaid(bb, units, RaiderReconChan)

	fmt.Println("entering range over RaiderRecon channel")
	for RaiderRecon := range RaiderReconChan {
		if RaiderRecon.Error != nil {
			log.Println(RaiderRecon.Error)
			continue
		}

		fmt.Println(RaiderRecon)

		//calls to marshaling the order data and submiting order to Oanda
		//need to send over r.RaiderStatus and OrderID
		if RaiderRecon.ExecuteOrder == 1 {
			//sending current create/submit order code over channel
			BBStatusChan <- RaiderRecon
		}
			RaiderRecon.Orders.OrderData.Units = units

			//creating []byte order data for the order HTTP body
			ordersByte := oanda.MarshalOrders(RaiderRecon.Orders)

			//creating/submiting the order to oanda
			createOrderByte, err := oanda.CreateOrder(ordersByte)

			//checking CreateOrder error
			if err != nil {
				log.Println(err)
			}

			//unmarshaling the returned createOrderByte into a native struct
			orderCreateTransaction := oanda.OrderCreateTransaction{}.
				UnmarshalOrderCreateTransaction(createOrderByte)

			//accessing the orderID field and saving it to a variable
			orderID := orderCreateTransaction.OrderFillTransaction.OrderID


			r.OrderID = orderID

		}
		wg.Wait()
	}

}

//CheckOrder used an OrderID to get the latest order status
func (r *Raider) CheckOrder() {
	//using the orderID to check the order status
	for {
		if r.RaiderStatus == 1 {
			checkOrderByte, err := oanda.CheckOrder(r.OrderID)
			if err != nil{
				fmt.Println(err)
			}
			fmt.Println("check order byte:")
			fmt.Println(checkOrderByte)
		}
	}
}

//RaiderRecon contains order data and the execute order decision
type RaiderRecon struct {
	Orders       oanda.Orders
	ExecuteOrder int //1 = execute order. 0 = don't execute order
	Error        error
}

//SingleRaid compares a single PricesData to a BollingerBand and returns a trading decision
func (r RaiderRecon) SingleRaid(bb BollingerBand, units string) RaiderRecon {
	//initializing pricesData struct
	pricesData := PricesData{}.Init(bb.Instrument)

	if pricesData.Error != nil {
		return RaiderRecon{Error: pricesData.Error}
	}

	//calling CalculateSpread also sets bid/ask fields in struct
	//pricesData.CalculateSpread()

	//FIXME need to have better error handling here
	if pricesData.Bid > bb.UpperBand {
		return RaiderRecon{
			Orders: oanda.Orders{}.MarketSellOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 1,
		}
	} else if pricesData.Ask < bb.LowerBand {
		return RaiderRecon{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 1,
		}
	} else {
		return RaiderRecon{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 0,
		}
	}

}

//ContinuousRaid ranges over a channel of PricesData comparing each PricesData to a
//BollingerBand and sends a trading decision over a channel to the caller
//FIXME if running this function coninuosly be careful to generate a new
//bollinger band at the start of each day
func (r RaiderRecon) ContinuousRaid(bb *BollingerBand, units string, out chan RaiderRecon) {
	oandaChan := make(chan PricesData)
	go StreamBidAsk(bb.Instrument, oandaChan)

	for pricesData := range oandaChan {
		if pricesData.Error != nil {
			out <- RaiderRecon{Error: pricesData.Error}
		}

		if pricesData.Heartbeat != nil {
			fmt.Println(pricesData.Heartbeat)
			continue
		}

		//print prices data
		//fmt.Println(pricesData)

		//FIXME need to have better error handling here
		if pricesData.Bid > bb.UpperBand {
			out <- RaiderRecon{
				Orders: oanda.Orders{}.MarketSellOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				ExecuteOrder: 1,
			}
		} else if pricesData.Ask < bb.LowerBand {
			out <- RaiderRecon{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				ExecuteOrder: 1,
			}
		} else {
			out <- RaiderRecon{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				ExecuteOrder: 0,
			}
		}
	}
}
