package prometheus

import (
	"errors"
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

func (client *PromClient) query(endpoint string) (string, error) {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := httpClient.Get(client.address + "/api/v1/query?query=" + endpoint)
	if err != nil {
		log.Printf("HTTP Error")
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Could not read prometheus response body")
		return "", err
	}
	bodyString := string(body)
	if !gjson.Valid(bodyString) {
		log.Println("Received unvalid json from prometheus")
		return "", errors.New("received Invalid json from Prometheus")
	}
	if success != gjson.Get(bodyString, "status").String() {
		log.Println("Failed")
		return "", nil
	}

	return bodyString, nil
}

func (client *PromClient) Headers() int64 {
	response, err := client.query("bitcoin_headers")
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	queryResult := gjson.Get(response, "data.result").Array()
	firstResultValue := queryResult[0].Map()["value"].Array()
	log.Println("Headers: ", firstResultValue[1].Int())
	return firstResultValue[1].Int()
}

func (client *PromClient) Blocks() int64 {
	response, err := client.query("bitcoin_blocks")
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	queryResult := gjson.Get(response, "data.result").Array()
	firstResultValue := queryResult[0].Map()["value"].Array()
	log.Println("Blocks: ", firstResultValue[1].Int())
	return firstResultValue[1].Int()
}

func (client *PromClient) VerificationProgress() float64 {
	response, err := client.query("bitcoin_verification_progress")
	if err != nil {
		log.Println(err.Error())
		return 0.0
	}
	queryResult := gjson.Get(response, "data.result").Array()
	firstResultValue := queryResult[0].Map()["value"].Array()
	log.Println("Verification Progress: ", firstResultValue[1].Float())
	return firstResultValue[1].Float()
}
