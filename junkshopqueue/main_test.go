package main

import (
	"fmt"
	"testing"
)

const(
	buyMsg = "i will buy it"
	failedBuyMsg = "sold"
	successfullyBuyMsg = "worth it :))"
	whichOneMsg = "which one?"

	//info
	infoMsg = "what do dou have?"
	nothingMsg = "i have nothing"

	//sell
	thxMsg = "thank you"
)

type client struct {
	ch chan string
}

func (c *client) buy() string {
	c.ch <- buyMsg
	answer := <-c.ch
	return answer
}

func (c *client) sell(name string) string {
	c.ch <- fmt.Sprintf("i have %s", name)
	answer := <-c.ch
	return answer
}

func (c *client) getInfo() string {
	c.ch <- infoMsg
	answer := <-c.ch
	return answer
}

func newClient(shopChannel chan chan string) client {
	cl := client{
		ch: make(chan string),
	}
	shopChannel <- cl.ch
	return cl
}

func TestSimple(t *testing.T) {
	shopChannel := make(chan chan string)
	go ManageShop(shopChannel)
	client_1 := newClient(shopChannel)

	product_1 := "scenery"

	if buyRes := client_1.buy(); buyRes != whichOneMsg {
		t.Errorf("your answer must be: %s but found: %s", whichOneMsg, buyRes)
	}

	if infoRes := client_1.getInfo(); infoRes != nothingMsg {
		t.Errorf("your answer must be: %s but found: %s", nothingMsg, infoRes)
	}

	if sellRes := client_1.sell(product_1); sellRes != thxMsg {
		t.Errorf("your answer must be: %s but found: %s", thxMsg, sellRes)
	}

	if buyRes := client_1.buy(); buyRes != whichOneMsg {
		t.Errorf("your answer must be: %s but found: %s", whichOneMsg, buyRes)
	}

	if infoRes := client_1.getInfo(); infoRes != product_1 {
		t.Errorf("your answer must be: %s but found: %s", product_1, infoRes)
	}

	if buyRes := client_1.buy(); buyRes != successfullyBuyMsg {
		t.Errorf("your answer must be: %s but found: %s", successfullyBuyMsg, buyRes)
	}

	
	if buyRes := client_1.buy(); buyRes != failedBuyMsg {
		t.Errorf("your answer must be: %s but found: %s", failedBuyMsg, buyRes)
	}

	fmt.Printf("Test successfully passed")
}