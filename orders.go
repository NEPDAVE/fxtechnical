package fxtechnical

import (
	//"log"
	//"strconv"
	"errors"
	oanda "github.com/nepdave/oanda"
)

//FIXME the error handling in this file is not correct! Your git pull never worked!!!!!!
//FIXME think about creating a CreateBuyOrder and CreateSellOrder func. This will
//make things more readable. Also think about error handling and possibly another
//function that will prepare the order and another function that will execute it
//also take a look at the data structure and make sure it's getting marshalled
//correctly also add a func to Unmarshal the data after placing an order
//total side note also lookinto coding the double bb
func CreateBuyOrder(bid float64, ask float64, instrument string, units int) Orders {
	//FIXME stopLossPrice and takeProfitPrice are hardcoded to certain ratios
	//this may not be the best way... infact it may be good to determine these
	//rations in the fxtechnical package instead
	targetPrice := bid
	stopLossPrice := bid - .00002
	takeProfitPrice := bid + (ask - bid) - .000002
	stopLoss := StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}
	takeProfit := TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}
	orderData := Order{
		Price:            targetPrice,
		StopLossOnFill:   stopLoss,
		TakeProfitOnFill: takeProfit,
		TimeInForce:      "FOK",
		Instrument:       instrument,
		Type:             "LIMIT",
		PositionFill:     "DEFAULT"}
	order := OrderBody{Order: orderData}

	jsonOrders, err := json.Marshal(order)
	if err != nil {
		log.Printf("Json Marshal Error: %s\n", err)
	}

	return jsonOrders

}

//FIXME remember that this units param should be negative
func CreateSellOrder(bid float64, ask float64, instrument string, units int) Orders {
	//FIXME stopLossPrice and takeProfitPrice are hardcoded to certain ratios
	//this may not be the best way...
	targetPrice = ask
	stopLossPrice = ask + .00002
	takeProfitPrice = bid - (ask - bid) + .000002
	stopLoss := StopLossOnFill{TimeInForce: "GTC", Price: stopLossPrice}
	takeProfit := TakeProfitOnFill{TimeInForce: "GTC", Price: takeProfitPrice}
	orderData := Order{
		Price:            targetPrice,
		StopLossOnFill:   stopLoss,
		TakeProfitOnFill: takeProfit,
		TimeInForce:      "FOK",
		Instrument:       instrument,
		Type:             "LIMIT",
		PositionFill:     "DEFAULT"}
	order := OrderBody{Order: orderData}

	jsonOrders, err := json.Marshal(order)
	if err != nil {
		log.Printf("Json Marshal Error: %s\n", err)
	}

	return jsonOrders

}
