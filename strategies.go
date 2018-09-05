package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"sync"
	"time"
)

/*
***************************
Raider is a trading algorithm that implements the Bolinger Band indicator
***************************
*/

//Raider holds the algo state and neccesary algo data
type Raider struct {
	Instrument      string
	mu              sync.Mutex
	OrderState      string       //closed/pending/open.
	CreateOrderCode int          //0 = dont execute 1 = execute
	OrderID         string       "none" //OrderID of current order
	Orders          oanda.Orders //Order SL/TP Limit/Market data
	Util            OrderUtilities
	Error           error
}

/*
General flow
PrepareOrder()
CreateOrder()
GetOrderID()
CheckOrder()
*/

//Init kicks off the select pattern to create orders and check orders
func (r Raider) Init(instrument string, units string) {
	var wg sync.WaitGroup

	RaiderChan := make(chan Raider)
	r.Instrument = instrument

	wg.Add(2)
	go r.ExecuteContinuosRaid(instrument, units, RaiderChan)
	go r.ExecuteContinuousGetOrderStatus()

	wg.Wait()
}

//ExecuteContinuosRaid ranges over the RaiderChan to execute orders and update
//Raider fields
func (r *Raider) ExecuteContinuosRaid(instrument string, units string, RaiderChan chan Raider) {
	var wg sync.WaitGroup

	wg.Add(1)
	//Checks whether or not conditions are right to trade
	go r.ContinuousRaid(instrument, RaiderChan)

	for raider := range RaiderChan {
		if raider.Error != nil {
			log.Println(raider.Error)
			continue
		}

		if raider.CreateOrderCode == 1 && r.OrderState == "closed" {
			fmt.Println("received create order signal...")
			r.mu.Lock()
			//doing exspensive IO calls but need to verify OrderState
			r.OrderID = r.Util.ExecuteOrder(instrument, units, raider.Orders)
			r.OrderState = r.Util.GetOrderStatus(r.OrderID)
			r.mu.Unlock()
		} else {
			fmt.Printf("Create Order Code = %d\n", raider.CreateOrderCode)
		}
	}
	wg.Wait()
}

//ExecuteContinuousGetOrderStatus ranges over the GetOrderStatusChan to update
//the Raider Status field
func (r *Raider) ExecuteContinuousGetOrderStatus() {
	for {
		r.mu.Lock()
		r.OrderState = r.Util.GetOrderStatus(r.OrderID)
		r.mu.Unlock()
		fmt.Println("")
		fmt.Printf("ORDER-ID %s %s STATE = %s\n", r.OrderID, r.Instrument, r.OrderState)
	}
}

//SingleRaid compares a single PricesData to a BollingerBand and returns Orders
//and the CreateOrderCode. 1 = execute order, 0 = don't execute order
func (r *Raider) SingleRaid(instrument string) (oanda.Orders, int) {
	bb := BollingerBand{}.Init(instrument, "20", "D")
	pricesData := PricesData{}.Init(instrument)

	if pricesData.Error != nil {
		log.Println(pricesData.Error)
		return oanda.Orders{}, 0
	}

	//setting all units to 0 here so that proper amount of units can be set later
	if pricesData.Bid > bb.UpperBand {
		return oanda.Orders{}.MarketSellOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 1
	} else if pricesData.Ask < bb.LowerBand {
		return oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 1
	} else {
		return oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 0
	}
}

//ContinuousRaid ranges over a channel of PricesData comparing each PricesData to a
//BollingerBand and sends a trading decision over a channel to the caller
func (r *Raider) ContinuousRaid(instrument string, RaiderChan chan Raider) {
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

	oandaChan := make(chan PricesData)
	go StreamBidAsk(bb.Instrument, oandaChan)

	for pricesData := range oandaChan {
		if pricesData.Error != nil {
			RaiderChan <- Raider{Error: pricesData.Error}
		}

		if pricesData.Heartbeat != nil {
			fmt.Println(pricesData.Heartbeat)
			continue
		}

		if pricesData.Bid > bb.UpperBand {
			RaiderChan <- Raider{
				Orders: oanda.Orders{}.MarketSellOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument, "0"),
				CreateOrderCode: 1,
			}
		} else if pricesData.Ask < bb.LowerBand {
			RaiderChan <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument, "0"),
				CreateOrderCode: 1,
			}
		} else {
			RaiderChan <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument, "0"),
				CreateOrderCode: 0,
			}
		}
	}
}
