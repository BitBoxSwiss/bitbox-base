package prometheus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

const host = "localhost"

// Prometheus is an interface representing a prometheus query client
type Prometheus interface {
	QueryFloat64(string) (float64, error)
}

// Client can query values from prometheus
type Client struct {
	port string
}

// NewClient returns a new prometheus client.
// It does not ensure that the client has connectivity.
func NewClient(port string) (client Client) {
	client.port = port
	return
}

// queryJSON queries prometheus with the specified expression and returns the JSON as a string
func (c Client) queryJSON(expression string) (string, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	url := "http://" + host + ":" + c.port + "/api/v1/query?query=" + expression
	httpResp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to perform a get GET on the prometheus server: %s", err.Error())
	}
	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from prometheus request: %s", err.Error())
	}

	if !gjson.Valid(string(body)) { // check if the response is valid json
		return "", fmt.Errorf("prometheus request returned invalid JSON: %s", string(body))
	}

	return string(body), nil
}

// parsePrometheusResponseAsFloat parses a prometheus JSON response and returns a float
func (c Client) parsePrometheusResponseAsFloat(json string) (float64, error) {

	// Check for a valid prometheus json response by checking:
	// - the `status` == success
	// - the list `data.result` having one and only one entry
	// - the value list `data.result[0].value` having exactly two entires
	// - there exists a response value for our expression `data.result[0].value[1]`

	status := gjson.Get(json, "status").String()
	if status != "success" {
		return -1, fmt.Errorf("prometheus request returned non-success (%s): %v", status, json)
	}

	queryResult := gjson.Get(json, "data.result").Array()
	if len(queryResult) != 1 {
		return -1, fmt.Errorf("unexpectedly got %d results from prometheus request: %s", len(queryResult), json)
	}

	firstResultValue := queryResult[0].Map()["value"].Array()
	if len(firstResultValue) != 2 {
		return -1, fmt.Errorf("unexpectedly got %d values from prometheus request: %s", len(firstResultValue), json)
	}

	if firstResultValue[1].Exists() == false {
		return -1, fmt.Errorf("the result value does not exist: %s", json)
	}

	measuredValue := firstResultValue[1].Float()
	return measuredValue, nil
}

// QueryFloat64 querys a float64 value for a given expression
func (c Client) QueryFloat64(expression string) (val float64, err error) {
	json, err := c.queryJSON(expression)
	if err != nil {
		return 0, fmt.Errorf("Could not query %s from prometheus: %s", expression, err.Error())
	}

	val, err = c.parsePrometheusResponseAsFloat(json)
	if err != nil {
		return 0, fmt.Errorf("Could parse response from '%s' query: %s", expression, err.Error())
	}
	return val, err
}
