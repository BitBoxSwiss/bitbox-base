package lightning

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tidwall/gjson"
)

var InvoiceListeningTimeout = time.Minute * 150
var WaitSendPayTimeout = time.Hour * 24
var WaitPaymentMaxAttempts = 60

type Client struct {
	Path             string
	PaymentHandler   func(gjson.Result)
	LastInvoiceIndex int
}

// ListenForInvoices starts a goroutine that will repeatedly call waitanyinvoice.
// Each payment received will be fed into the client.PaymentHandler function.
// You can change that function in the meantime.
// Or you can set it to nil if you want to stop listening for invoices.
func (ln *Client) ListenForInvoices() {
	go func() {
		for {
			if ln.PaymentHandler == nil {
				log.Print("won't listen for invoices: no PaymentHandler.")
				return
			}

			res, err := ln.CallWithCustomTimeout(InvoiceListeningTimeout,
				"waitanyinvoice", ln.LastInvoiceIndex)
			if err != nil {
				if _, ok := err.(ErrorTimeout); ok {
					time.Sleep(time.Minute)
				} else {
					log.Printf("error waiting for invoice %d: %s", ln.LastInvoiceIndex, err.Error())
					time.Sleep(5 * time.Second)
				}
				continue
			}

			index := res.Get("pay_index").Int()
			ln.LastInvoiceIndex = int(index)

			ln.PaymentHandler(res)
		}
	}()
}

// PayAndWaitUntilResolution implements its 'pay' logic, querying and retrying routes.
// It's like the default 'pay' plugin, but it blocks until a final success or failure is achieved.
// After it returns you can be sure a failed payment will not succeed anymore.
// Any value in params will be passed to 'getroute' or 'sendpay' or smart defaults will be used.
// This includes values from the default 'pay' plugin.
func (ln *Client) PayAndWaitUntilResolution(
	bolt11 string,
	params map[string]interface{},
) (success bool, payment gjson.Result, tries []Try, err error) {
	decoded, err := ln.Call("decodepay", bolt11)
	if err != nil {
		return false, payment, tries, err
	}

	hash := decoded.Get("payment_hash").String()
	fakePayment := gjson.Parse(`{"payment_hash": "` + hash + `"}`)

	exclude := []string{}
	payee := decoded.Get("payee").String()
	delayFinalHop := decoded.Get("min_final_cltv_expiry").Int()

	var msatoshi float64
	if imsatoshi, ok := params["msatoshi"]; ok {
		if converted, err := toFloat(imsatoshi); err == nil {
			msatoshi = converted
		}
	} else {
		msatoshi = decoded.Get("msatoshi").Float()
	}

	riskfactor, ok := params["riskfactor"]
	if !ok {
		riskfactor = 10
	}
	label, ok := params["label"]
	if !ok {
		label = ""
	}

	maxfeepercent := 0.5
	if imaxfeepercent, ok := params["maxfeepercent"]; ok {
		if converted, err := toFloat(imaxfeepercent); err == nil {
			maxfeepercent = converted
		}
	}
	exemptfee := 5000.0
	if iexemptfee, ok := params["exemptfee"]; ok {
		if converted, err := toFloat(iexemptfee); err == nil {
			exemptfee = converted
		}
	}

	routehints := decoded.Get("routes").Array()

	if len(routehints) > 0 {
		for _, rh := range routehints {
			done, payment := tryPayment(ln, &tries, bolt11,
				payee, msatoshi, hash, label, &exclude,
				delayFinalHop, riskfactor, maxfeepercent, exemptfee, &rh)
			if done {
				return true, payment, tries, nil
			}
		}
	} else {
		done, payment := tryPayment(ln, &tries, bolt11,
			payee, msatoshi, hash, label, &exclude,
			delayFinalHop, riskfactor, maxfeepercent, exemptfee, nil)
		if done {
			return true, payment, tries, nil
		}
	}

	return false, fakePayment, tries, nil
}

func tryPayment(
	ln *Client,
	tries *[]Try,
	bolt11 string,
	payee string,
	msatoshi float64,
	hash string,
	label interface{},
	exclude *[]string,
	delayFinalHop int64,
	riskfactor interface{},
	maxfeepercent float64,
	exemptfee float64,
	hint *gjson.Result,
) (paid bool, payment gjson.Result) {
	for try := 0; try < 30; try++ {
		target := payee
		if hint != nil {
			target = hint.Get("0.pubkey").String()
		}

		res, err := ln.CallNamed("getroute",
			"id", target,
			"riskfactor", riskfactor,
			"cltv", delayFinalHop,
			"msatoshi", msatoshi,
			"fuzzpercent", 0,
			"exclude", *exclude,
		)
		if err != nil {
			// no route or invalid parameters, call it a simple failure
			return
		}

		if !res.Get("route").Exists() {
			continue
		}

		route := res.Get("route")
		// if we're using a route hint, increment the queried route with the missing parts
		if hint != nil {
			route = addHintToRoute(route, *hint, payee, delayFinalHop)
		}

		// inspect route, it shouldn't be too expensive
		if route.Get("0.msatoshi").Float()/msatoshi > (1 + 1/maxfeepercent) {
			// too expensive, but we'll still accept it if the payment is small
			if msatoshi > exemptfee {
				// otherwise try the next route
				// we force that by excluding a channel
				*exclude = append(*exclude, getWorstChannel(route))
				continue
			}
		}

		// ignore returned value here as we'll get it from waitsendpay below
		_, err = ln.CallNamed("sendpay",
			"route", route.Value(),
			"payment_hash", hash,
			"label", label,
			"bolt11", bolt11,
		)
		if err != nil {
			// the command may return an error and we don't care
			if _, ok := err.(ErrorCommand); ok {
				// we don't care because we'll see this in the next call
			} else {
				// otherwise it's a different odd error, stop
				return
			}
		}

		// this should wait indefinitely, but 24h is enough
		payment, err = ln.CallWithCustomTimeout(WaitSendPayTimeout, "waitsendpay", hash)
		if err != nil {
			if cmderr, ok := err.(ErrorCommand); ok {
				*tries = append(*tries, Try{route.Value(), &cmderr, false})

				switch cmderr.Code {
				case 200, 202:
					// try again
					continue
				case 204:
					// error in route, eliminate erring channel and try again
					data, ok0 := cmderr.Data.(map[string]interface{})
					ichannel, ok1 := data["erring_channel"]
					channel, ok2 := ichannel.(string)

					if !ok0 || !ok1 || !ok2 {
						// should never happen
						return
					}

					// if erring channel is in the route hint just stop altogether
					if hint != nil {
						for _, hhop := range hint.Array() {
							if hhop.Get("short_channel_id").String() == channel {
								return
							}
						}
					}

					// get erring channel a direction by inspecting the route
					var direction int64
					for _, hop := range route.Array() {
						if hop.Get("channel").String() == channel {
							direction = hop.Get("direction").Int()
							goto gotdirection
						}
					}

					// we should never get here
					return

				gotdirection:
					*exclude = append(*exclude, fmt.Sprintf("%s/%d", channel, direction))
					continue
				}
			}

			// a different error, call it a complete failure
			return
		}

		// payment suceeded
		*tries = append(*tries, Try{route.Value(), nil, true})
		return true, payment
	}

	// stop trying
	return
}

func getWorstChannel(route gjson.Result) (worstChannel string) {
	var worstFee int64 = 0
	hops := route.Array()
	if len(hops) == 1 {
		return hops[0].Get("channel").String() + "/" + hops[0].Get("direction").String()
	}

	for i := 0; i+1 < len(hops); i++ {
		hop := hops[i]
		next := hops[i+1]
		fee := hop.Get("msatoshi").Int() - next.Get("msatoshi").Int()
		if fee > worstFee {
			worstFee = fee
			worstChannel = hop.Get("channel").String() + "/" + hop.Get("direction").String()
		}
	}

	return
}

func addHintToRoute(
	route gjson.Result, hint gjson.Result,
	finalPeer string, finalHopDelay int64,
) gjson.Result {
	var extrafees int64 = 0  // these extra fees will be added to the public part
	var extradelay int64 = 0 // this extra delay will be added to the public part

	// we know exactly the length of our new route
	npublichops := route.Get("#").Int()
	nhinthops := hint.Get("#").Int()
	newroute := make([]map[string]interface{}, npublichops+nhinthops)

	// so we can start adding the last hops (from the last and backwards)
	r := len(newroute) - 1

	lastPublicHop := route.Array()[npublichops-1]

	hhops := hint.Array()
	for h := len(hhops) - 1; h >= 0; h-- {
		hhop := hhops[h]

		nextdelay, delaydelta, nextmsat, fees, nextpeer := grabParameters(
			hint,
			newroute,
			lastPublicHop,
			finalPeer,
			finalHopDelay,
			r,
			h,
		)

		// delay for this hop is anything in the next hop plus the delta
		delay := nextdelay + delaydelta

		// calculate this channel direction
		var direction int
		if hhop.Get("pubkey").String() < nextpeer {
			direction = 1
		} else {
			direction = 0
		}

		newroute[r] = map[string]interface{}{
			"id":        nextpeer,
			"channel":   hhop.Get("short_channel_id").Value(),
			"direction": direction,
			"msatoshi":  int64(nextmsat) + fees,
			"delay":     delay,
		}

		// bump extra stuff for the public part
		extrafees += fees
		extradelay += delaydelta

		r--
	}

	// since these parameters are always based on the 'next' part of the route, we need
	// to run a fake thing here with the hint channel at index -1 so we'll get the parameters
	// for actually index 0 -- this is not to add them to the actual route, but only to
	// grab the 'extra' fees/delay we need to apply to the public part of the route
	_, delaydelta, _, fees, _ := grabParameters(
		hint,
		newroute,
		lastPublicHop,
		finalPeer,
		finalHopDelay,
		r,
		-1,
	)
	extrafees += fees
	extradelay += delaydelta
	// ~

	// now we start from the beggining with the public part of the route
	r = 0
	route.ForEach(func(_, hop gjson.Result) bool {
		newroute[r] = map[string]interface{}{
			"id":        hop.Get("id").Value(),
			"channel":   hop.Get("channel").Value(),
			"direction": hop.Get("direction").Value(),
			"delay":     hop.Get("delay").Int() + extradelay,
			"msatoshi":  hop.Get("msatoshi").Int() + extrafees,
		}
		r++
		return true
	})

	// turn it into a gjson.Result
	newroutejsonstr, _ := json.Marshal(newroute)
	newroutegjson := gjson.ParseBytes(newroutejsonstr)

	return newroutegjson
}

func grabParameters(
	fullHint gjson.Result,
	fullNewRoute []map[string]interface{},
	lastPublicHop gjson.Result,
	finalPeer string,
	finalHopDelay int64,
	r int, // the full route hop index we're working on
	h int, // the hint part channel index we're working on
) (
	nextdelay int64, // delay amount for the hop after this or the final node's cltv
	delaydelta int64, // delaydelta is given by the next hop hint or 0
	nextmsat int64, // msatoshi amount for the hop after this (or the final amount)
	fees int64, // fees are zero in the last hop, or a crazy calculation otherwise
	nextpeer string, // next node id (or the final node)
) {
	if int64(h) == fullHint.Get("#").Int()-1 {
		// this is the first iteration, means it's the last hint channel/hop
		nextmsat = lastPublicHop.Get("msatoshi").Int() // this is the final amount, yes it is.
		nextdelay = finalHopDelay
		nextpeer = finalPeer
		delaydelta = 0
		fees = 0
	} else {
		// now we'll get the value of a hop we've just calculated/iterated over
		nextHintHop := fullNewRoute[r+1]
		nextmsat = nextHintHop["msatoshi"].(int64)
		nextdelay = nextHintHop["delay"].(int64)

		nextHintChannel := fullHint.Array()[h+1]
		nextpeer = nextHintChannel.Get("pubkey").String()
		delaydelta = nextHintChannel.Get("cltv_expiry_delta").Int()

		// fees for this hop are based on the next
		fees = nextHintChannel.Get("fee_base_msat").Int() +
			int64(
				(float64(nextmsat)/1000)*nextHintChannel.Get("fee_proportional_millionths").Float()/1000,
			)

	}

	return
}

type Try struct {
	Route   interface{}   `json:"route"`
	Error   *ErrorCommand `json:"error"`
	Success bool          `json:"success"`
}
