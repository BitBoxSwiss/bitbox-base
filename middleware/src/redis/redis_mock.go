package redis

import (
	"strconv"
)

// MockClient is a mock redis client
type MockClient struct {
	mockRedisMap map[string]string
}

// NewMockClient returns a new redis client.
// It does not ensure that the client has connectivity.
func NewMockClient(port string) (mockClient *MockClient) {
	mockRedisMap := make(map[string]string)
	mockRedisMap["base:version"] = "0.0.1"
	return &MockClient{mockRedisMap: mockRedisMap}
}

// SetString sets a mock value to given key.
func (mc *MockClient) SetString(key string, value string) error {
	mc.mockRedisMap[key] = value
	return nil
}

// GetInt gets an integer value for a given key.
func (mc *MockClient) GetInt(key string) (val int, err error) {
	s := mc.mockRedisMap[key]
	val, err = strconv.Atoi(s)
	return
}

// GetString gets an string for a given key.
func (mc *MockClient) GetString(key string) (val string, err error) {
	return mc.mockRedisMap[key], nil
}
