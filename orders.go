package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
)

/*
***************************
Functions to create and prepare Simple Orders
***************************
*/

//CreateSimpleOrder creates the an order with Oanda that does not contain a Stop Loss
func CreateSimpleOrder(units string, instrument string, takeProfit string) (*oanda.OrderCreateTransactionResponse, error) {

	//building market order struct
	orders := SimpleMarketOrders(units, instrument, takeProfit)

	//marshaling market order struct into byte slice
	ordersByte := oanda.SimpleOrdersRequest{}.MarshalSimpleOrdersRequest(orders)

	fmt.Println(string(ordersByte))

	//posting market order byte slice to oanda api
	response, err := oanda.CreateOrder(ordersByte)

	fmt.Println(string(response))

	if err != nil {
		logger.Println(err)
		return &oanda.OrderCreateTransactionResponse{}, err
	}

	//unmarshaling order create transaction into struct
	transaction, err := oanda.OrderCreateTransactionResponse{}.
		UnmarshalOrderCreateTransactionResponse(response)

	return transaction, err
}

//MarketOrder builds struct needed for marshaling data into a []byte
func SimpleMarketOrders(units string, instrument string, takeProfitPrice string) oanda.SimpleOrdersRequest {

	//FIXME need to figure out how to create trailing stop loss
	//trailingStopLossOnFill := oanda.TrailingStopLossOnFill{}

	//take profit data
	takeProfitOnFill := oanda.TakeProfitOnFill{
		TimeInForce: "GTC", Price: takeProfitPrice,
	}

	//order data
	orders := oanda.SimpleOrdersRequest{
		Orders: oanda.SimpleOrders{
			TakeProfitOnFill: takeProfitOnFill,
			TimeInForce:      "FOK",
			Instrument:       instrument,
			Units:            units,
			Type:             "MARKET",
			PositionFill:     "DEFAULT"},
	}

	return orders
}
