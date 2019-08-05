package prometheus

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

const (
	success = "success"
)

type PromClient struct {
	address string
}

func NewPromClient(address string) *PromClient {
	return &PromClient{
		address: address,
	}
}

func (client *PromClient) query(endpoint string) string {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := httpClient.Get(client.address + "/api/v1/query?query=" + endpoint)
	if err != nil {
		log.Printf("Some weird http error: %v", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Could not read prometheus response body")
	}
	bodyString := string(body)
	if !gjson.Valid(bodyString) {
		log.Println("Received unvalid json from prometheus")
	}

	return bodyString
}

func (client *PromClient) Headers() int64 {
	response := client.query("bitcoin_headers")
	if success != gjson.Get(response, "status").String() {
		log.Println("Failed")
	}
	queryResult := gjson.Get(response, "data.result").Array()
	firstResultValue := queryResult[0].Map()["value"].Array()
	log.Println("result")
	return firstResultValue[1].Int()
}

func (client *PromClient) Blocks() int64 {
	response := client.query("bitcoin_blocks")
	if success != gjson.Get(response, "status").String() {
		log.Println("Failed")
	}
	queryResult := gjson.Get(response, "data.result").Array()
	firstResultValue := queryResult[0].Map()["value"].Array()
	log.Println("result")
	return firstResultValue[1].Int()
}

func (client *PromClient) VerificationProgress() float64 {
	response := client.query("bitcoin_verification_progress")
	if success != gjson.Get(response, "status").String() {
		log.Println("Failed")
	}
	queryResult := gjson.Get(response, "data.result").Array()
	firstResultValue := queryResult[0].Map()["value"].Array()
	log.Println("result")
	return firstResultValue[1].Float()
}
