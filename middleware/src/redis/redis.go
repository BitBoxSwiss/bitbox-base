package redis

import (
	"fmt"
	"log"

	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/gomodule/redigo/redis"
)

// Redis is an interface representing a redis Client
type Redis interface {
	ConvertErrorToErrorResponse(error) rpcmessages.ErrorResponse
	GetBool(key BaseRedisKey) (bool, error)
	GetInt(BaseRedisKey) (int, error)
	GetString(BaseRedisKey) (string, error)
	SetString(BaseRedisKey, string) error
}

// Client is a redis client
type Client struct {
	pool *redis.Pool
}

// NewClient returns a new redis client.
// It does not ensure that the client has connectivity.
func NewClient(port string) (client Client) {
	pool := newPool(port)

	err := ping(pool.Get())
	if err != nil {
		// If the Redis server is not reachable on middleware start up the
		// supervisor should take over and restart (i.e. fix) the Redis server.
		log.Printf("Warning redis server connectivity could not be established: %s", err.Error())
	}
	return Client{pool: pool}
}

func newPool(port string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost"+":"+port)
			if err != nil {
				return nil, err
			}
			return conn, err
		},
	}
}

func ping(c redis.Conn) (err error) {
	defer c.Close()
	_, err = c.Do("PING")
	if err != nil {
		return
	}
	return
}

// getConnection gets a connection from the pool
func (c Client) getConnection() redis.Conn {
	return c.pool.Get()
}

// GetInt gets an integer value for a given key.
func (c Client) GetInt(key BaseRedisKey) (val int, err error) {
	conn := c.getConnection()
	val, err = redis.Int(conn.Do("GET", key))
	if err != nil {
		return -1, fmt.Errorf("could not get key %s as integer: %s", key, err.Error())
	}
	return val, nil
}

// GetBool gets a boolean value for a given key.
// Internally checks if the value for the given key is set to 1.
// If so, then true is returned, else false.
func (c Client) GetBool(key BaseRedisKey) (val bool, err error) {
	conn := c.getConnection()
	valAsInt, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		return false, fmt.Errorf("could not get key %s as boolean: %s", key, err.Error())
	}
	return valAsInt == 1, nil
}

// GetString gets a string for a given key.
func (c Client) GetString(key BaseRedisKey) (val string, err error) {
	conn := c.getConnection()
	val, err = redis.String(conn.Do("GET", key))
	if err != nil {
		return "", fmt.Errorf("could not get key %s as string: %s", key, err.Error())
	}
	return val, nil
}

// SetString sets a string for a given key.
func (c Client) SetString(key BaseRedisKey, value string) error {
	conn := c.getConnection()
	_, err := conn.Do("SET", key, value)
	if err != nil {
		return fmt.Errorf("could not set key %s: %s", key, err.Error())
	}
	return nil
}

// ConvertErrorToErrorResponse converts an error returned by Redis to an ErrorResponse
func (c Client) ConvertErrorToErrorResponse(err error) rpcmessages.ErrorResponse {
	return rpcmessages.ErrorResponse{
		Success: false,
		Message: err.Error(),
		Code:    rpcmessages.ErrorRedisError,
	}
}
