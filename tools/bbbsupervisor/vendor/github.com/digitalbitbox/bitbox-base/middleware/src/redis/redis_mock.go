package redis

import (
	"log"
	"strconv"

	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
)

// MockClient is a mock redis client
type MockClient struct {
	mockRedisMap map[string]string
}

// NewMockClient returns a new redis client.
// It does not ensure that the client has connectivity.
func NewMockClient(port string) (mockClient *MockClient) {
	mockRedisMap := setupTestData()
	return &MockClient{mockRedisMap: mockRedisMap}
}

// SetString sets a mock value to given key.
func (mc *MockClient) SetString(key BaseRedisKey, value string) error {
	mc.mockRedisMap[string(key)] = value
	return nil
}

// GetInt gets an integer value for a given key.
func (mc *MockClient) GetInt(key BaseRedisKey) (val int, err error) {
	s := mc.mockRedisMap[string(key)]
	val, err = strconv.Atoi(s)
	return
}

// GetBool gets a boolean value for a given key.
// Internally checks if the value for the given key is set to 1.
// If so, then true is returned, else false.
// GetInt gets an integer value for a given key.
func (mc *MockClient) GetBool(key BaseRedisKey) (val bool, err error) {
	s := mc.mockRedisMap[string(key)]
	valAsInt, err := strconv.Atoi(s)
	return valAsInt == 1, err
}

// AddToSortedSet is a dummy that does nothing but printing the arguments.
func (mc *MockClient) AddToSortedSet(key BaseRedisKey, score int, element string) error {
	log.Printf("AddToSortedSet dummy: key %q, score %q, element %q", key, score, element)
	return nil
}

// RemoveFromSortedSet is a dummy that does nothing but printing the arguments.
func (mc *MockClient) RemoveFromSortedSet(key BaseRedisKey, element string) error {
	log.Printf("RemoveFromSortedSet dummy: key %q, element %q", key, element)
	return nil
}

// GetTopFromSortedSet is a dummy that does nothing but printing the arguments.
func (mc *MockClient) GetTopFromSortedSet(key BaseRedisKey) (string, error) {
	log.Printf("GetTopFromSortedSet dummy: key %q", key)
	return "dummy mock", nil
}

// GetString gets an string for a given key.
func (mc *MockClient) GetString(key BaseRedisKey) (val string, err error) {
	return mc.mockRedisMap[string(key)], nil
}

func setupTestData() map[string]string {
	mockRedisMap := make(map[string]string)

	// General mock data
	mockRedisMap[string(BaseVersion)] = "0.0.1"
	mockRedisMap[string(BaseHostname)] = "bitbox-base-redis-mock"
	mockRedisMap[string(TorEnabled)] = "1"
	mockRedisMap[string(BitcoindListen)] = "1"
	mockRedisMap[string(MiddlewarePasswordSet)] = "0"
	mockRedisMap[string(BaseSetupDone)] = "0"
	mockRedisMap[string(MiddlewareAuth)] = `{"admin":{"password":"ICanHasPasword?","role":"admin"}}`

	// Specific test values for testing util.go getBooleanFromRedis()
	// TestGetBooleanFromRedis() in util_test.go
	mockRedisMap["test:getBooleanFromRedis:true"] = "1"
	mockRedisMap["test:getBooleanFromRedis:false1"] = "0"
	mockRedisMap["test:getBooleanFromRedis:false2"] = "3"
	mockRedisMap["test:getBooleanFromRedis:false3"] = "abc"

	// Specific test values for testing util.go getStringFromRedis()
	// TestGetStringFromRedis() in util_test.go
	mockRedisMap["test:getStringFromRedis:abc"] = "abc"
	mockRedisMap["test:getStringFromRedis:empty"] = ""

	return mockRedisMap
}

// ConvertErrorToErrorResponse converts an error returned by Redis to an ErrorResponse
func (mc *MockClient) ConvertErrorToErrorResponse(err error) rpcmessages.ErrorResponse {
	return rpcmessages.ErrorResponse{
		Success: false,
		Message: err.Error(),
		Code:    rpcmessages.ErrorRedisError,
	}
}
