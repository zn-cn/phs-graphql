package db

import (
	"config"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

// redis controller
type RedisDBCntlr struct {
	conn redis.Conn
}

var globalRedisPool *redis.Pool
var redisURL string
var redisPW string

func init() {
	redisConf := config.Conf.Redis
	redisURL = fmt.Sprintf("%s:%s", redisConf.Host, redisConf.Port)
	redisPW = redisConf.PW
	globalRedisPool = GetRedisPool()
}

// GetRedisPool get the client pool of redis
func GetRedisPool() *redis.Pool {
	pool := &redis.Pool{ // 实例化一个连接池
		MaxIdle:     30, // 最大的连接数量
		MaxActive:   0,  // 连接池最大连接数量,不确定可以用0（0表示自动定义）
		IdleTimeout: 60, // 连接关闭时间 60秒 （60秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { // 要连接的redis数据库
			conn, err := redis.Dial("tcp", redisURL)
			if err != nil {
				return nil, err
			}
			if redisPW != "" {
				_, err = conn.Do("AUTH", redisPW)
			}
			return conn, err
		},
	}
	return pool
}

/********************************************* RedisDBCntlr *******************************************/

func NewRedisDBCntlr() *RedisDBCntlr {
	return &RedisDBCntlr{
		conn: globalRedisPool.Get(),
	}
}

func (this *RedisDBCntlr) Close() {
	this.conn.Close()
}

func (this *RedisDBCntlr) GetConn() redis.Conn {
	return this.conn
}

func (this *RedisDBCntlr) Send(commandName string, args ...interface{}) error {
	return this.conn.Send(commandName, args...)
}

func (this *RedisDBCntlr) Do(commandName string, args ...interface{}) (interface{}, error) {
	return this.conn.Do(commandName, args...)
}

func (this *RedisDBCntlr) GET(key string) (string, error) {
	return redis.String(this.conn.Do("GET", key))
}

func (this *RedisDBCntlr) GETInt64(key string) (int64, error) {
	return redis.Int64(this.conn.Do("GET", key))
}

func (this *RedisDBCntlr) SET(key string, value interface{}) (interface{}, error) {
	return this.conn.Do("SET", key, value)
}

func (this *RedisDBCntlr) SETEX(key string, expire int64, value string) (interface{}, error) {
	return this.conn.Do("SETEX", key, expire, value)
}

func (this *RedisDBCntlr) INCRBY(key string, num int) (interface{}, error) {
	return this.conn.Do("INCRBY", key, num)
}

func (this *RedisDBCntlr) KEYS(keyPattern string) ([]string, error) {
	return redis.Strings(this.conn.Do("KEYS", keyPattern))
}

func (this *RedisDBCntlr) DEL(keys ...interface{}) (interface{}, error) {
	return this.conn.Do("DEL", keys...)
}

func (this *RedisDBCntlr) HGETALL(key string) (map[string]interface{}, error) {
	d, err := redis.StringMap(this.conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	for k, v := range d {
		data[k] = v
	}
	return data, err
}

func (this *RedisDBCntlr) HMGET(key string, fields ...interface{}) (map[string]string, error) {
	args := []interface{}{key}
	args = append(args, fields...)
	return redis.StringMap(this.conn.Do("HMGET", args...))
}

func (this *RedisDBCntlr) HMSET(key string, fields ...interface{}) (interface{}, error) {
	args := []interface{}{key}
	args = append(args, fields...)
	return this.conn.Do("HMSET", args...)
}

func (this *RedisDBCntlr) LRANGE(key string, start, end int) ([]string, error) {
	return redis.Strings(this.conn.Do("LRANGE", key, start, end))
}

func (this *RedisDBCntlr) LLEN(key string) (int, error) {
	return redis.Int(this.conn.Do("LLEN", key))
}

func (this *RedisDBCntlr) LPOP(key string) (string, error) {
	return redis.String(this.conn.Do("LPOP", key))
}

func (this *RedisDBCntlr) RPUSH(key string, params ...interface{}) (interface{}, error) {
	args := []interface{}{key}
	args = append(args, params...)
	return this.conn.Do("RPUSH", args...)
}

func (this RedisDBCntlr) EXPIRE(key string, seconds int64) (interface{}, error) {
	return this.conn.Do("EXPIRE", key, seconds)
}
