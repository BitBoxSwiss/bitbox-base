Useful and fast interface for [lightningd](https://github.com/ElementsProject/lightning/). All methods return [gjson.Result](https://godoc.org/github.com/tidwall/gjson#Result), which is a good thing (unless your application relies on a lot of information from many different lightningd responses) and you can [learn it in 10 seconds](https://github.com/tidwall/gjson#get-a-value).

[![godoc.org](https://img.shields.io/badge/reference-godoc-blue.svg)](https://godoc.org/github.com/fiatjaf/lightningd-gjson-rpc)

This is a simple and resistant client. It comes with a practical **invoice listener** method and nice (could be nicer?) defaults for retrying and connecting to faulty lightningd nodes.

### Usage

```go
package main

import (
  "github.com/fiatjaf/lightningd-gjson-rpc"
  "github.com/tidwall/gjson"
)

var ln *lightning.Client

func main () {
    lastinvoiceindex := getFromSomewhereOrStartAtZero()

    ln = &lightning.Client{
        Path:             "/home/whatever/.lightning/lightning-rpc",
        LastInvoiceIndex: lastinvoiceindex, // only needed if you're going to listen for invoices
        PaymentHandler:   handleInvoicePaid, // only needed if you're going to listen for invoices
    }
    ln.ListenForInvoices() // optional

    nodeinfo, err := ln.Call("getinfo")
    if err != nil {
        log.Fatal("getinfo error: " + err.Error())
    }

    log.Print(nodeinfo.Get("alias").String())
}

// this is called with the result of `waitanyinvoice`
func handlePaymentReceived(inv gjson.Result) {
    index := inv.Get("pay_index").Int()
    saveSomewhere(index)

    hash := inv.Get("payment_hash").String()
    log.Print("one of our invoices was paid: " + hash)
}
```

### Passing parameters

There are three modes of passing parameters, you can call either:

```go
// 1. `Call` with a list of parameters, in the order defined by each command;
ln.Call("invoice", 1000000, "my-label", "my description", 3600)

// 2. `Call` with a single `map[string]interface{}` with all parameters properly named; or
ln.Call("invoice", map[string]interface{
    "msatoshi": "1000000,
    "label": "my-label",
    "description": "my description",
    "preimage": "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
})

// 3. `CallNamed` with a list of keys and values passed in the proper order.
ln.CallNamed("invoice",
    "msatoshi", "1000000,
    "label", "my-label",
    "description", "my description",
    "preimage", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
    "expiry", 3600,
)
```
