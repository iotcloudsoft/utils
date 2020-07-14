package xredis

import (
	"encoding/json"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Engine struct {
	pool *redis.Pool
}

type Setting struct {
	Host        string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

// Setup Initialize the Redis instance
func NewEngine(setting *Setting) (*Engine, error) {
	pool := &redis.Pool{
		MaxIdle:     setting.MaxIdle,
		MaxActive:   setting.MaxActive,
		IdleTimeout: setting.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", setting.Host)
			if err != nil {
				return nil, err
			}
			if setting.Password != "" {
				if _, err := c.Do("AUTH", setting.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &Engine{pool}, nil
}

// Exists check a key
func (e *Engine) Exists(key string) bool {
	conn := e.pool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false
	}

	return exists
}

// GetObject get a key
func (e *Engine) GetObject(key string, data interface{}) error {
	conn := e.pool.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return err
	}

	err = json.Unmarshal(reply, data)
	if err != nil {
		return err
	}

	return nil
}

// HGetObject get a hash key
func (e *Engine) HGetObject(key string, field string, data interface{}) error {
	conn := e.pool.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("HGET", key, field))
	if err != nil {
		return err
	}

	err = json.Unmarshal(reply, data)
	if err != nil {
		return err
	}

	return nil
}

// HGetMapString get map string
func (e *Engine) HGetMapString(key string) (map[string]string, error) {
	conn := e.pool.Get()
	defer conn.Close()

	reply, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// SetObject a key/value
func (e *Engine) SetObject(key string, data interface{}) error {
	conn := e.pool.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, value)
	if err != nil {
		return err
	}

	return nil
}

// HSetObject get a hash key
func (e *Engine) HSetObject(key string, field string, data interface{}) error {
	conn := e.pool.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", key, field, value)
	if err != nil {
		return err
	}

	return nil
}

// Delete delete a kye
func (e *Engine) Delete(key string) (bool, error) {
	conn := e.pool.Get()
	defer conn.Close()

	return redis.Bool(conn.Do("DEL", key))
}

// LikeDeletes batch delete
func (e *Engine) LikeDeletes(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", "*"+key+"*"))
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err = conn.Do("DEL", key)
		if err != nil {
			return err
		}
	}

	return nil
}
