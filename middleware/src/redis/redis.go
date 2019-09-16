package redis

import (
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

// Client represents a redis client that has a connection pool.
type Client struct {
	pool *redis.Pool
}

// NewRedisClient returns a new redis client.
// It does not ensure that the client has connectivity.
func NewRedisClient(port string) (client *Client) {
	pool := newPool(port)

	err := ping(pool.Get())
	if err != nil {
		// If the Redis server is not reachable on middleware start up the
		// supervisor should take over and restart (i.e. fix) the Redis server.
		log.Printf("Warning redis server connectivity could not be established: %s", err.Error())
	}
	client = &Client{pool: pool}
	return
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
func (c *Client) getConnection() redis.Conn {
	return c.pool.Get()
}

// GetInt gets an integer value for a given key.
func (c *Client) GetInt(key string) (val int, err error) {
	conn := c.getConnection()
	val, err = redis.Int(conn.Do("GET", key))
	if err != nil {
		return -1, fmt.Errorf("could not key %s as integer: %s", key, err.Error())
	}
	return val, nil
}

// GetString gets an string for a given key.
func (c *Client) GetString(key string) (val string, err error) {
	conn := c.getConnection()
	val, err = redis.String(conn.Do("GET", key))
	if err != nil {
		return "", fmt.Errorf("could not get key %s as string: %s", key, err.Error())
	}
	return val, nil
}
