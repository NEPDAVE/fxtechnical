package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
)

var (
	logger *log.Logger
)

func FxTechInit(logger *log.Logger) {
	logger = logger
}

/*
***************************
Collection of functions to create orders with oanda
***************************
*/

//CreateOrder creates the order data structure, submits the order to Oanda,
//and returns an OrderID
func CreateOrder(units string, instrument string, stopLoss string, takeProfit string) (*oanda.OrderCreateTransactionResponse, error) {


	//building market order struct
	orders := MarketOrder(units, instrument, stopLoss, takeProfit)

	//marshaling market order struct into byte slice
	ordersByte := oanda.OrdersRequest{}.MarshalOrdersRequest(orders)

	//posting market order byte slice to oanda api
	response, err := oanda.CreateOrder(ordersByte)

	if err != nil {
		logger.Println(err)
		return &oanda.OrderCreateTransactionResponse{}, err
	}

	//unmarshaling order create transaction into struct
	transaction, err := oanda.OrderCreateTransactionResponse{}.
		UnmarshalOrderCreateTransactionResponse(response)

	return transaction, err
}

/*
***************************
Collection of functions to prepare orders
***************************
*/

//MarketOrder builds struct needed for marshaling data into a []byte
func MarketOrder(units string, instrument string, stopLossPrice string,
	takeProfitPrice string) oanda.OrdersRequest {

	//stop loss data
	stopLossOnFill := oanda.StopLossOnFill{
		TimeInForce: "GTC", Price: stopLossPrice,
	}

	//FIXME need to figure out how to create trailing stop loss
	//trailingStopLossOnFill := oanda.TrailingStopLossOnFill{}

	//take profit data
	takeProfitOnFill := oanda.TakeProfitOnFill{
		TimeInForce: "GTC", Price: takeProfitPrice,
	}

	//submiting order with no takeProfit or stoploss
	if stopLossPrice == "" && takeProfitPrice == "" {
		//order data
		orders := oanda.OrdersRequest{
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
	orders := oanda.OrdersRequest{
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
	units string) oanda.OrdersRequest {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", (targetPrice - .002))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", (targetPrice + .006))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	stringTargetPrice := fmt.Sprintf("%.5f", targetPrice)

	orders := oanda.OrdersRequest{
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
	units string) oanda.OrdersRequest {
	//tp/sl ratio is 3 to 1
	stopLossPrice := fmt.Sprintf("%.5f", (targetPrice + .002))
	stopLossOnFill := oanda.StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}

	takeProfitPrice := fmt.Sprintf("%.5f", (targetPrice - .006))
	takeProfitOnFill := oanda.TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}

	stringTargetPrice := fmt.Sprintf("%.5f", targetPrice)

	orders := oanda.OrdersRequest{
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
func TrailingStopLossOrder(tradeID string, distance string) oanda.OrdersRequest {
	orders := oanda.OrdersRequest{
		Orders: oanda.Orders{
			Type:             "TRAILING_STOP_LOSS",
			TradeID:          tradeID,
			Distance:         distance,
			TimeInForce:      "GTC",
			TriggerCondition: "DEFAULT"},
	}

	return orders
}
