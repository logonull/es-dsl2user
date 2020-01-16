package models

import (
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisPool *redis.Pool

func init() {
	address := beego.AppConfig.String("redisAddress")
	password := beego.AppConfig.String("redisPassword")
	dbNum := beego.AppConfig.String("redisDb")

	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", address)
		if err != nil {
			return nil, err
		}

		if beego.AppConfig.String("redisPassword") != "" {
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
		}

		_, selecterr := c.Do("SELECT", dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}
	// initialize a new pool
	RedisPool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   30,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
	}
	RedisPool.Wait = true
}
