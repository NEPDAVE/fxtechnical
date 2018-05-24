package fxtechnical

import (
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

//TestBidAsk validates BidAsk returns two strings
func TestGetPricing(t *testing.T) {
	bid, ask := BidAsk("EUR_USD")

	t.Log("Given the need to test fxtechnical API wrapper")
	t.Log("\tWhen checking BidAsk function")

	if bid != "0" && ask != "0"  {
		t.Fatal("\t\tShould return two non zero strings", ballotX)
	}
	t.Log("\t\tShould return two non zero strings", checkMark)
}
