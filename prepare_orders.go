package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
)

/*
***************************
Collection of functions to create/submit Market and Limit orders to oanda
***************************
*/

//this shit need to return an oanda.OrderCreateTransaction
func CreateOrder(units string, instrument string, stopLoss string, takeProfit string) {
	clientOrder := MarketOrder(units, instrument, stopLoss, takeProfit)

	clientOrderByte := oanda.ClientOrders{}.MarshalClientOrders(clientOrder)

	respByte, err := oanda.CreateOrder(clientOrderByte)

	if err != nil{
		fmt.Println(err)
	}

	//FIXME this looks weird/bad
	fmt.Println(string(respByte))
	fmt.Println(err)
}

/*
***************************
Collection of functions to prepare Market and Limit orders
***************************
*/

//MarketOrder builds struct needed for marshaling data into a []byte
func MarketOrder(stopLossPrice string, takeProfitPrice string,
	instrument string, units string) oanda.ClientOrders {

	//stop loss data
	stopLossOnFill := oanda.StopLossOnFill{
		TimeInForce: "GTC", Price: stopLossPrice,
	}


	// //trailing stop loss data
	// trailingStopLossOnFill := oanda.TrailingStopLossOnFill{}

	//take profit data
	takeProfitOnFill := oanda.TakeProfitOnFill{
		TimeInForce: "GTC", Price: takeProfitPrice,
	}

	//submiting order with no takeProfit or stoploss
	if stopLossPrice == "" && takeProfitPrice == "" {
		//order data
		orders := oanda.ClientOrders{
			Orders: oanda.Orders{
				TimeInForce:  "FOK",
				Instrument:   instrument,
				Units:        units,
				Type:         "MARKET",
				PositionFill: "DEFAULT"},
		}

		fmt.Println("test me - line 45, prepare_orders.go")

		return orders

	}

	//order data
	orders := oanda.ClientOrders{
		Orders: oanda.Orders{
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "FOK",
			Instrument:       instrument,
			Units:            units,
			Type:             "MARKET",
			PositionFill:     "DEFAULT"},
	}

	return orders
}

//LimitLongOrder builds struct needed for marshaling data into a []byte
//FIXME SL/TP is hard coded. need to do more research here
func LimitLongOrder(targetPrice float64, instrument string,
	units string) oanda.ClientOrders {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", (targetPrice - .002))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", (targetPrice + .006))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	stringTargetPrice := fmt.Sprintf("%.5f", targetPrice)

	orders := oanda.ClientOrders{
		Orders: oanda.Orders{
			Price:            stringTargetPrice,
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "GTC",
			Instrument:       instrument,
			Type:             "LIMIT",
			PositionFill:     "DEFAULT"},
	}

	return orders
}

//LimitShortOrder builds struct needed for marshaling data into a []byte
//FIXME SL/TP is hard coded. need to do more research here
func LimitShortOrder(targetPrice float64, instrument string,
	units string) oanda.ClientOrders {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", (targetPrice + .002))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", (targetPrice - .006))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	stringTargetPrice := fmt.Sprintf("%.5f", targetPrice)

	orders := oanda.ClientOrders{
		Orders: oanda.Orders{
			Price:            stringTargetPrice,
			StopLossOnFill:   stopLossOnFill,
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "GTC",
			Instrument:       instrument,
			Type:             "LIMIT",
			PositionFill:     "DEFAULT"},
	}

	return orders
}

/*
***************************
Trailing Stop Loss Order
***************************
*/

//TrailingStopLossOrder builds struct needed for marshaling data into a []byte
func TrailingStopLossOrder(tradeID string, distance string) oanda.ClientOrders {
	orders := oanda.ClientOrders{
		Orders: oanda.Orders{
			Type:             "TRAILING_STOP_LOSS",
			TradeID:          tradeID,
			Distance:         distance,
			TimeInForce:      "GTC",
			TriggerCondition: "DEFAULT"},
	}

	return orders
}

/*
{
  "order": {
    "price": "1.5000",
    "stopLossOnFill": {
      "timeInForce": "GTC",
      "price": "1.7000"
    },
    "takeProfitOnFill": {
      "price": "1.14530"
    },
    "timeInForce": "GTC",
    "instrument": "USD_CAD",
    "units": "-1000",
    "type": "LIMIT",
    "positionFill": "DEFAULT"
  }
	{{1.30100 {GTC 1.30300} {GTC 1.29500} GTC GBP_USD  LIMIT DEFAULT}}
*/
