package fxtechnical

import (
	oanda "github.com/nepdave/oanda"
)

/*
***************************
Raider is a trading algorithm that implements the the Bolinger Band indicator
***************************
*/

//Raider contains order data and the execute order decision
type Raider struct {
	Orders       oanda.Orders
	ExecuteOrder int //1 = execute order. 0 = don't execute order
	Error        error
}

//SingleRaid compares a single PricesData to a BollingerBand and returns a trading decision
func (r Raider) SingleRaid(bb BollingerBand, units int) Raider {
	//initializing pricesData struct
	pricesData := PricesData{}.Init(bb.Instrument)

	if pricesData.Error != nil {
		return Raider{Error: pricesData.Error}
	}

	//calling CalculateSpread also sets bid/ask fields in struct
	pricesData.CalculateSpread()

	//FIXME need to have better error handling here
	if pricesData.Bid > bb.UpperBand {
		return Raider{
			Orders: oanda.Orders{}.MarketSellOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 1,
		}
	} else if pricesData.Ask < bb.LowerBand {
		return Raider{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 1,
		}
	} else {
		return Raider{
			ExecuteOrder: 0,
		}
	}

}

// //ContinuousRaid ranges over a channel of PricesData comparing each PricesData to a
// //BollingerBand and sends a trading decision over a channel to the caller
// func (r Raider) ContinuousRaid(bb BollingerBand, units int) {
// 	//initializing pricesData struct
// 	pricesData := PricesData{}.Init(bb.Instrument)
//
// 	if pricesData.Error != nil {
// 		//log.Println(pricesData.Error)
// 	}
//
// 	//calling CalculateSpread also sets bid/ask fields in struct
// 	pricesData.CalculateSpread()
//
// 	if pricesData.Bid > bb.UpperBand {
// 		//returning Orders struct along with a 1. 1 = submit order
// 		oanda.Orders{}.MarketSellOrder(bid, ask, instrument, units), 1
// 	} else if pricesData.Ask < bb.LowerBand {
// 		//returning Orders struct along with a 1. 1 = submit order
// 		return oanda.Orders{}.MarketBuyOrder(bid, ask, instrument, units), 1
// 	} else {
// 		//returning empty Orders struct along with a 0. 0 = don't submit order
// 		return oanda.Orders{}, 0
// 	}
//
// }
