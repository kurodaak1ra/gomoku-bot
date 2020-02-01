package redis

import (
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

var redisAddr string = "127.0.0.1:6379"

// SET
func SET(arr [][]string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	for _, val := range arr {
		c.Send("SET", val[0], val[1])
		if val[2] != "-1" {
			c.Send("EXPIRE", val[0], val[2])
		}
	}
	c.Flush()
	s, err := c.Receive()
	// defer c.Close()
	return s, err
}

// LSET
func LSET(name string, index, value int) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("LSET", name, index, value)
	c.Flush()
	ls, err := c.Receive()
	// defer c.Close()
	return ls, err
}

// LLEN
func LLEN(key string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("LLEN", key)
	c.Flush()
	ll, err := c.Receive()
	// defer c.Close()
	return ll, err
}

// LPUSH
func LPUSH(key string, val string, timeout int) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("LPUSH", key, val)
	c.Send("EXPIRE", key, timeout)
	c.Flush()
	lp, err := c.Receive()
	// defer c.Close()
	return lp, err
}

func LPOP(key string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("LPOP", key)
	c.Flush()
	lp, err := c.Receive()
	// defer c.Close()
	if lp == nil {
		return nil, err
	} else {
		return string(lp.([]uint8)), err
	}
}

// LRANGE
func LRANGE(key string, start, end int) ([]string, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("LRANGE", key, start, end)
	c.Flush()
	lp, err := c.Receive()
	// defer c.Close()
	var tempArr []string
	for _, val := range lp.([]interface{}) {
		tempArr = append(tempArr, string(val.([]uint8)))
	}
	return tempArr, err
}

// DEL
func DEL(arr []string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	for _, val := range arr {
		c.Send("DEL", val)
	}
	c.Flush()
	d, err := c.Receive()
	// defer c.Close()
	return d, err
}

// DEL BATTLE（没有批量删除，只能 for 来实现）
func CLEAR(chatID int64, uid int) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	// 遍历
	c.Send("KEYS", strconv.FormatInt(chatID, 10)+":battle-*"+strconv.Itoa(uid)+"*")
	c.Flush()
	db, _ := c.Receive()
	// 没有数据表示有人在搞事
	if len(db.([]interface{})) == 0 {
		return nil, nil
	}
	// 删除
	for _, val := range db.([]interface{}) {
		c.Send("DEL", string(val.([]uint8)))
	}
	c.Flush()
	d, err := c.Receive()
	// defer c.Close()
	return d, err
}

// GET
func GET(key string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("GET", key)
	c.Flush()
	g, err := c.Receive()
	// defer c.Close()
	if g == nil {
		return nil, err
	} else {
		return string(g.([]uint8)), err
	}
}

// EXISTS
func EXISTS(key string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("EXISTS", key)
	c.Flush()
	e, err := c.Receive()
	// defer c.Close()
	return e, err
}

// GET KEYS
func KEYS(path string) ([]string, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("KEYS", path)
	c.Flush()
	g, err := c.Receive()
	// defer c.Close()
	var tempArr []string
	for _, val := range g.([]interface{}) {
		tempArr = append(tempArr, string(val.([]uint8)))
	}
	return tempArr, err
}

// GET TTL
func TTL(key string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	c.Send("TTL", key)
	c.Flush()
	t, err := c.Receive()
	// defer c.Close()
	return t, err
}

// SET EXPIRE
func EXPIRE(arr [][]string) (interface{}, error) {
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(fmt.Sprintln("Connect to redis error", err))
	}
	for _, val := range arr {
		c.Send("EXPIRE", val[0], val[1])
	}
	c.Flush()
	s, err := c.Receive()
	// defer c.Close()
	return s, err
}
