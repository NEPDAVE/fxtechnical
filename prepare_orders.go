package fxtechnical

import (
	//"log"
	//"strconv"
	//"errors"
	"fmt"
	oanda "github.com/nepdave/oanda"
)

/*
***************************
Collection of functions to prepare Market and Limit orders for creation
***************************
*/

//MarketLongOrder builds struct needed for marshaling data into a []byte
//FIXME should not be setting SL/TP in this package
func MarketLongOrder(bid float64, ask float64, instrument string,
	units string) oanda.Orders {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", bid-(ask*.005))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", ask+(ask*.015))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	orders := oanda.Orders{
		Order: oanda.Order{
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "FOK",
			Instrument:       instrument,
			Type:             "MARKET",
			PositionFill:     "DEFAULT"},
	}

	return orders
}

//MarketShortOrder builds struct needed for marshaling data into a []byte
//FIXME should not be setting SL/TP in this package
//FIXME sell orders units need to be a string with a -(negative) sign in front
func MarketShortOrder(bid float64, ask float64, instrument string,
	units string) oanda.Orders {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", ask+(bid*.005))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", bid-(ask*.015))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	orders := oanda.Orders{
		Order: oanda.Order{
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "FOK",
			Instrument:       instrument,
			Type:             "MARKET",
			PositionFill:     "DEFAULT"},
	}

	return orders
}

//LimitLongOrder builds struct needed for marshaling data into a []byte
//FIXME SL/TP is hard coded. need to do more research here
func LimitLongOrder(targetPrice float64, instrument string,
	units string) oanda.Orders {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", (targetPrice - .002))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", (targetPrice + .006))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	stringTargetPrice := fmt.Sprintf("%.5f", targetPrice)

	orders := oanda.Orders{
		Order: oanda.Order{
			Price:            stringTargetPrice,
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "GTC",
			Instrument:       instrument,
			Type:             "LIMIT",
			PositionFill:     "DEFAULT"},
	}

	fmt.Println(stringTargetPrice)

	return orders
}

//LimitShortOrder builds struct needed for marshaling data into a []byte
//FIXME SL/TP is hard coded. need to do more research here
func LimitShortOrder(targetPrice float64, instrument string,
	units string) oanda.Orders {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", (targetPrice + .002))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", (targetPrice - .006))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	stringTargetPrice := fmt.Sprintf("%.5f", targetPrice)

	orders := oanda.Orders{
		Order: oanda.Order{
			Price:            stringTargetPrice,
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "GTC",
			Instrument:       instrument,
			Type:             "LIMIT",
			PositionFill:     "DEFAULT"},
	}

	fmt.Println(stringTargetPrice)

	return orders
}