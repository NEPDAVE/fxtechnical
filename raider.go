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
***************************
Raider is a trading algorithm that implements the Bolinger Band indicator
***************************
*/

//Raider holds the algo state and neccesary algo data
type Raider struct {
	Instrument      string
	OrderState      string             //closed/pending/open.
	CreateOrderCode int                //0 = dont execute 1 = execute
	OrderID         string             //OrderID of current order
	Orders          oanda.ClientOrders //Order SL/TP Limit/Market data
	Error           error
}

//Init kicks off the goroutines to create orders and check orders
func (r Raider) Init(instrument string, units string) {
	var wg sync.WaitGroup

	RaiderChan := make(chan Raider)
	r.Instrument = instrument

	wg.Add(2)
	go r.ExecuteContinuosRaid(instrument, units, RaiderChan)
	go r.ExecuteContinuousGetOrder()

	wg.Wait()
}

//ExecuteContinuosRaid ranges over the RaiderChan to execute orders and update
//Raider fields
func (r *Raider) ExecuteContinuosRaid(instrument string, units string, RaiderChan chan Raider) {
	var wg sync.WaitGroup
	var mu sync.Mutex

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
			mu.Lock()
			//doing exspensive IO calls but need to verify OrderState
			r.OrderID = CreateClientOrdersAndGetOrderID(instrument, units,
				raider.Orders)
			r.OrderState = GetOrderState(r.OrderID)
			mu.Unlock()
		} else {
			fmt.Printf("Create Order Code = %d\n", raider.CreateOrderCode)
		}
	}
	wg.Wait()
}

//ExecuteContinuousGetOrder ranges over the GetOrderChan to update
//the Raider Status field
func (r *Raider) ExecuteContinuousGetOrder() {
	var mu sync.Mutex

	for {
		mu.Lock()
		r.OrderState = GetOrderState(r.OrderID)
		mu.Unlock()
		fmt.Println("")
		fmt.Printf("ORDER-ID %s %s STATE = %s\n", r.OrderID, r.Instrument, r.OrderState)
	}
}

//SingleRaid compares a single PricesData to a BollingerBand and returns Orders
//and the CreateOrderCode. 1 = execute order, 0 = don't execute order
func (r *Raider) SingleRaid(instrument string) (oanda.ClientOrders, int) {
	bb := BollingerBand{}.Init(instrument, "20", "D")
	pricesData := PricesData{}.Init(instrument)

	if pricesData.Error != nil {
		log.Println(pricesData.Error)
		return oanda.ClientOrders{}, 0
	}

	//setting all units to 0 here so that proper amount of units can be set later
	if pricesData.Bid > bb.UpperBand {
		return MarketShortOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 1
	} else if pricesData.Ask < bb.LowerBand {
		return MarketLongOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 1
	} else {
		return MarketLongOrder(pricesData.Bid,
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
			orders := MarketShortOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument, "0")
			RaiderChan <- Raider{Orders: orders, CreateOrderCode: 1}
		} else if pricesData.Ask < bb.LowerBand {
			orders := MarketLongOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument, "0")
			RaiderChan <- Raider{Orders: orders, CreateOrderCode: 1}
		} else {
			orders := MarketLongOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument, "0")
			RaiderChan <- Raider{Orders: orders, CreateOrderCode: 0}
		}
	}
}