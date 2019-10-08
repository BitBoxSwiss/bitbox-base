package prometheus

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/tidwall/gjson"
)

const (
	success = "success"
)

// response represents a Prometheus response returned by the query() function
type response string

// resultType represents a Prometheus JSON result type
// The resultType is specified by the value of `data.resultType` in the JSON result
type resultType string

const (
	vector resultType = "vector"
)

// Client is a Prometheus client
type Client struct {
	address string
}

// NewClient returns a new Prometheus client.
// It does not ensure that the client has connectivity.
func NewClient(address string) Client {
	return Client{
		address: address,
	}
}

// query queries the Prometheus server and returns the response.
/* Dummy Prometheus JSON response with the resultType being "vector":
  {
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [
				{
					"metric": {
						"__name__": "base_system_info",
						<metric>: <value>,
					},
					"value": [
						<timestamp>,
						<value>
					]
				}
			]
		}
  }
*/
func (client *Client) query(query BasePrometheusQuery) (response, error) {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	escapedQuery := url.QueryEscape(string(query))
	queryURL := client.address + "/api/v1/query?query=" + escapedQuery

	resp, err := httpClient.Get(queryURL)
	if err != nil {
		return "", fmt.Errorf("a HTTP error occurred: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %s", err.Error())
	}
	bodyString := string(body)
	if !gjson.Valid(bodyString) {
		return "", fmt.Errorf("received invalid JSON from as result of the Prometheus query: %s", bodyString)
	}
	if success != gjson.Get(bodyString, "status").String() {
		return "", fmt.Errorf("the Prometheus query failed: %s", bodyString)
	}

	return response(bodyString), nil
}

// checkResultType compares the result type of a Prometheus query to a passed expected result type.
// If the result type of the response is not equal to the passed type then an error is returend.
func (r response) checkResultType(expected resultType) error {
	queryResultType := gjson.Get(string(r), "data.resultType")
	if !queryResultType.Exists() {
		return fmt.Errorf("the JSON response does not have a field 'data.resultType'")
	}
	if queryResultType.String() != string(expected) {
		return fmt.Errorf("the JSON response field 'data.resultType' does not match the expected value of '%s'. It's '%s'", string(expected), queryResultType.String())
	}
	return nil
}

// getFirstResultValue returns the value of the first result of a JSON response from Prometheus.
// The first result value is that is being queried for by when not quering for a information over a time range.
// This function expects a string representing a Prometheus JSON response with `data.resultType` equaling `"vector"`.
func (r response) getFirstResultValue() (value gjson.Result, err error) {
	err = r.checkResultType(vector)
	if err != nil {
		return value, err
	}

	queryResult := gjson.Get(string(r), "data.result")
	if !queryResult.Exists() {
		return value, fmt.Errorf("the query result does not have '%s'", "data.result")
	}
	queryResults := queryResult.Array()
	if len(queryResults) == 0 {
		return value, fmt.Errorf("the query result does not have any result entries")
	}
	queryResultValue := queryResults[0].Map()["value"]
	if !queryResultValue.Exists() {
		return value, fmt.Errorf("the query result does not have '%s'", "data.result[0]['values']")
	}
	queryResultValues := queryResultValue.Array()
	if len(queryResultValues) < 2 {
		return value, fmt.Errorf("the first query result value has less than two entries")
	}

	// queryResultValues is a parsed JSON array of [<timestamp>,<value>]
	value = queryResultValues[1]
	return value, nil
}

// getMetricsMap returns the metrics map that a Prometheus response can include.
// The metrics map includes for example string key value pairs and are sometime
// used to store non-integer values in Prometheus.
func (r response) getMetricsMap() (metricMap map[string]gjson.Result, err error) {
	err = r.checkResultType(vector)
	if err != nil {
		return metricMap, err
	}

	queryResult := gjson.Get(string(r), "data.result")
	if !queryResult.Exists() {
		return metricMap, fmt.Errorf("the query result does not have '%s'", "data.result")
	}
	queryResults := queryResult.Array()
	if len(queryResults) == 0 {
		return metricMap, fmt.Errorf("the query result does not have any result entries")
	}
	queryResultMetric := queryResults[0].Map()["metric"]
	if !queryResultMetric.Exists() {
		return metricMap, fmt.Errorf("the query result does not have '%s'", "data.result[0]['metric']")
	}

	metricMap = queryResultMetric.Map()
	return metricMap, nil
}

// GetFloat queries Prometheus with the provided query and returns an int64.
func (client *Client) GetFloat(query BasePrometheusQuery) (float64, error) {
	response, err := client.query(query)
	if err != nil {
		return 0, fmt.Errorf("could not query an integer for query '%s': %s", query, err.Error())
	}
	value, err := response.getFirstResultValue()
	if err != nil {
		return 0, fmt.Errorf("could not get the first result value from '%s': %s", response, err.Error())
	}

	return value.Float(), nil
}

// GetInt queries Prometheus with the provided query and returns an int64.
func (client *Client) GetInt(query BasePrometheusQuery) (int64, error) {
	response, err := client.query(query)
	if err != nil {
		return 0, fmt.Errorf("could not query an integer for query '%s': %s", query, err.Error())
	}

	value, err := response.getFirstResultValue()
	if err != nil {
		return 0, fmt.Errorf("could not get the first result value from '%s': %s", response, err.Error())
	}

	return value.Int(), nil
}

// GetMetricString gets a metric from a Prometheus query.
// Metrics are returned by Prometheus as extra information for the result.
func (client *Client) GetMetricString(query BasePrometheusQuery, metric string) (string, error) {
	response, err := client.query(query)
	if err != nil {
		return "", fmt.Errorf("could not query '%s' from Prometheus: %s", query, err.Error())
	}

	metricsMap, err := response.getMetricsMap()
	if err != nil {
		return "", fmt.Errorf("could not read the metrics map from Prometheus (response: %s): %s", string(response), err.Error())
	}

	value := metricsMap[metric].String()
	return value, nil
}

// ConvertErrorToErrorResponse converts an error returned by Prometheus to an ErrorResponse
func (client Client) ConvertErrorToErrorResponse(err error) rpcmessages.ErrorResponse {
	return rpcmessages.ErrorResponse{
		Success: false,
		Message: err.Error(),
		Code:    rpcmessages.ErrorPrometheusError,
	}
}

// Blocks returns the bitcoin block count from Prometheus
// Deprecated: is needed util middleware.GetVerificationProgress is replaced by GetServiceInfo or refactored to use GetInt itself.
func (client *Client) Blocks() int64 {
	log.Println("Calling deprecated function prometheus.Blocks()")
	blocks, err := client.GetInt(BitcoinBlockCount)
	if err != nil {
		log.Printf("Deprecated function Blocks(): %s", err.Error())
		return 0
	}
	return blocks
}

// Headers returns the bitcoin header count from Prometheus
// Deprecated: is needed util middleware.GetVerificationProgress is replaced by GetServiceInfo or refactored to use GetInt itself.
func (client *Client) Headers() int64 {
	log.Println("Calling deprecated function prometheus.Headers()")
	headers, err := client.GetInt(BitcoinHeaderCount)
	if err != nil {
		log.Printf("Deprecated function Headers(): %s", err.Error())
		return 0
	}
	return headers
}

// VerificationProgress returns the bitcoin verification progress from Prometheus
// Deprecated: is needed util middleware.GetVerificationProgress is replaced by GetServiceInfo or refactored to use GetInt itself.
func (client *Client) VerificationProgress() float64 {
	log.Println("Calling deprecated function prometheus.VerificationProgress()")
	verificationProgress, err := client.GetFloat(BitcoinVerificationProgress)
	if err != nil {
		log.Printf("Deprecated function VerificationProgress(): %s", err.Error())
		return 0
	}
	return verificationProgress
}
