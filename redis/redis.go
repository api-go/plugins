package redis

import (
	"encoding/json"
	"fmt"
	"github.com/api-go/plugin"
	"github.com/ssgo/log"
	"github.com/ssgo/redis"
	"github.com/ssgo/u"
)

type Redis struct {
	pool *redis.Redis
}

var redisPool = map[string]*redis.Redis{}
var defaultRedis *redis.Redis

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api-go/plugins/redis",
		Name: "redis",
		ConfigSet: []plugin.ConfigSet{
			{Name: "default", Type: "string", Memo: "默认的Redis连接，使用 redis.get() 来获得实例，格式为 redis://127.0.0.1:6379/1 或 redis://:<**加密的密码**>@127.0.0.1:6379?timeout=10s&logSlow=100ms"},
			{Name: "configs", Type: "map[string]string", Memo: "其他Redis连接，使用 redis.get('name') 来获得实例"},
		},
		Init: func(conf map[string]interface{}) {
			if conf["default"] != nil {
				defaultRedis = redis.GetRedis(u.String(conf["default"]), nil)
			}
			if conf["configs"] != nil {
				confs := map[string]string{}
				u.Convert(conf["configs"], &confs)
				for name, url := range confs {
					redisPool[name] = redis.GetRedis(url, nil)
				}
			}
		},
		Objects: map[string]interface{}{
			"fetch": GetRedis,
		},
	})
}

// GetRedis 获得Redis连接
// GetRedis name 连接配置名称，如果不提供名称则使用默认连接
// GetRedis return Redis连接，对象内置连接池操作，完成后无需手动关闭连接
func GetRedis(name *string, logger *log.Logger) *Redis {
	if name == nil || *name == "" {
		if defaultRedis != nil {
			return &Redis{pool: defaultRedis.CopyByLogger(logger)}
		}
	} else {
		if redisPool[*name] != nil {
			fmt.Println(u.JsonP(redisPool[*name].Config), 999)
			return &Redis{pool: redisPool[*name].CopyByLogger(logger)}
		} else if defaultRedis != nil {
			return &Redis{pool: defaultRedis.CopyByLogger(logger)}
		}
	}
	return &Redis{
		pool: redis.GetRedis("", logger),
	}
}

func makeRedisResult(r *redis.Result) interface{} {
	var v interface{}
	buf := r.Bytes()
	if json.Unmarshal(buf, &v) == nil {
		return v
	} else {
		return string(buf)
	}
}

func makeRedisResults(rr *redis.Result) []interface{} {
	out := make([]interface{}, 0)
	for _, r := range rr.Results() {
		out = append(out, makeRedisResult(&r))
	}
	return out
}

func makeRedisResultMap(rr *redis.Result) map[string]interface{} {
	out := map[string]interface{}{}
	for k, r := range rr.ResultMap() {
		out[k] = makeRedisResult(r)
	}
	return out
}

// Do 执行Redis操作，这是底层接口
// Do cmd 命令
// Do values 参数，根据命令传入不同参数
// Do return 返回字符串数据，需要根据数据内容自行解析处理
func (rd *Redis) Do(cmd string, values ...interface{}) string {
	return rd.pool.Do(cmd, values...).String()
}

// Del 删除
// Del keys 传入一个或多个Key
// Del return 成功删除的个数
func (rd *Redis) Del(keys ...string) int {
	return rd.pool.Do("DEL", u.ToInterfaceArray(keys)...).Int()
}

// Exists 判断是否Key存在
// * key 指定一个Key
// Exists return 是否存在
func (rd *Redis) Exists(key string) bool {
	return rd.pool.Do("EXISTS " + key).Bool()
}

// Expire 设置Key的过期时间
// * seconds 过期时间的秒数
// Expire return 是否成功
func (rd *Redis) Expire(key string, seconds int) bool {
	return rd.pool.Do("EXPIRE "+key, seconds).Bool()
}

// ExpireAt 设置Key的过期时间（指定具体时间）
// ExpireAt time 过期时间的时间戳，单位秒
// ExpireAt return 是否成功
func (rd *Redis) ExpireAt(key string, time int) bool {
	return rd.pool.Do("EXPIREAT "+key, time).Bool()
}

// Keys 查询Key
// * patten 查询条件，例如："SESS_*"
// * return []string 查询到的Key列表
func (rd *Redis) Keys(patten string) []string {
	return rd.pool.Do("KEYS " + patten).Strings()
}

// Get 读取Key的内容
// * return any 如果是一个对象则返回反序列化后的对象，否则返回字符串
func (rd *Redis) Get(key string) interface{} {
	return makeRedisResult(rd.pool.Do("GET " + key))
}

// GetEX 读取Key的内容并更新过期时间
func (rd *Redis) GetEX(key string, seconds int) interface{} {
	r := rd.pool.Do("GET " + key)
	if r.String() != "" {
		rd.pool.EXPIRE(key, seconds)
	}
	return makeRedisResult(r)
}

// Set 存储内容到Key
// * value 对象或字符串
// * return bool 是否成功
func (rd *Redis) Set(key string, value interface{}) bool {
	return rd.pool.Do("SET "+key, value).Bool()
}
// SetEX 存储内容到Key并设置过期时间
func (rd *Redis) SetEX(key string, seconds int, value interface{}) bool {
	return rd.pool.Do("SETEX "+key, seconds, value).Bool()
}
// SetNX 存储内容到一个不存在的Key，如果Key已经存在则设置失败
func (rd *Redis) SetNX(key string, value interface{}) bool {
	return rd.pool.Do("SETNX "+key, value).Bool()
}
// GetSet 将给定 key 的值设为 value ，并返回 key 的旧值(old value)
func (rd *Redis) GetSet(key string, value interface{}) interface{} {
	return makeRedisResult(rd.pool.Do("GETSET "+key, value))
}

// Incr 将 key 中储存的数值增一
// Incr * int64 最新的计数
func (rd *Redis) Incr(key string) int64 {
	return rd.pool.Do("INCR " + key).Int64()
}
// Decr 将 key 中储存的数值减一
func (rd *Redis) Decr(key string) int64 {
	return rd.pool.Do("DECR " + key).Int64()
}
// IncrBy 将 key 中储存的数值加上增量 increment
func (rd *Redis) IncrBy(key string, increment int64) int64 {
	return rd.pool.Do("INCRBY " + key, increment).Int64()
}
// DecrBy 将 key 中储存的数值减去加上增量 increment
func (rd *Redis) DecrBy(key string, increment int64) int64 {
	return rd.pool.Do("DECRBY " + key, increment).Int64()
}

// MGet 获取所有(一个或多个)给定 key 的值
// MGet return []any 按照查询key的顺序返回结果，如果结果是一个对象则返回反序列化后的对象，否则返回字符串
func (rd *Redis) MGet(keys ...string) []interface{} {
	return makeRedisResults(rd.pool.Do("MGET", u.ToInterfaceArray(keys)...))
}
// MSet 同时设置一个或多个 key-value 对
// MSet keyAndValues 按照key-value的顺序依次传入一个或多个数据
func (rd *Redis) MSet(keyAndValues ...interface{}) bool {
	return rd.pool.Do("MSET", keyAndValues...).Bool()
}

// HGet 获取存储在哈希表中指定字段的值
// * field 字段
func (rd *Redis) HGet(key, field string) interface{} {
	return makeRedisResult(rd.pool.Do("HGET "+key, field))
}
// HSet 将哈希表 key 中的字段 field 的值设为 value
func (rd *Redis) HSet(key, field string, value interface{}) bool {
	return rd.pool.Do("HSET "+key, field, value).Bool()
}

// HSetNX 只有在字段 field 不存在时，设置哈希表字段的值
func (rd *Redis) HSetNX(key, field string, value interface{}) bool {
	return rd.pool.Do("HSETNX "+key, field, value).Bool()
}
// HMGet 获取所有给定字段的值
// * fields 字段列表
func (rd *Redis) HMGet(key string, fields ...string) []interface{} {
	return makeRedisResults(rd.pool.Do("HMGET", append(append([]interface{}{}, key), u.ToInterfaceArray(fields)...)...))
}
// HGetAll 获取在哈希表中指定 key 的所有字段和值
// HGetAll return 返回所有字段的值，如果值是一个对象则返回反序列化后的对象，否则返回字符串
func (rd *Redis) HGetAll(key string) map[string]interface{} {
	return makeRedisResultMap(rd.pool.Do("HGETALL " + key))
}
// HMSet 将哈希表 key 中的字段 field 的值设为 value
func (rd *Redis) HMSet(key string, fieldAndValues ...interface{}) bool {
	return rd.pool.Do("HMSET", append(append([]interface{}{}, key), fieldAndValues...)...).Bool()
}
// HKeys 获取所有哈希表中的字段
func (rd *Redis) HKeys(patten string) []string {
	return rd.pool.Do("HKEYS " + patten).Strings()
}
// HLen 获取哈希表中字段的数量
// HLen return 字段数量
func (rd *Redis) HLen(key string) int {
	return rd.pool.Do("HLEN " + key).Int()
}
// HDel 删除一个或多个哈希表字段
// HDel return 成功删除的个数
func (rd *Redis) HDel(key string, fields ...string) int {
	return rd.pool.Do("HDEL", append(append([]interface{}{}, key), u.ToInterfaceArray(fields)...)...).Int()
}
// HExists 查看哈希表 key 中，指定的字段是否存在
func (rd *Redis) HExists(key, field string) bool {
	return rd.pool.Do("HEXISTS "+key, field).Bool()
}
// HIncr 为哈希表 key 中的指定字段的整数值加上增量1
func (rd *Redis) HIncr(key, field string) int64 {
	return rd.pool.Do("HINCRBY "+key, field, 1).Int64()
}
// HDecr 为哈希表 key 中的指定字段的整数值减去增量1
func (rd *Redis) HDecr(key, field string) int64 {
	return rd.pool.Do("HDECRBY "+key, field, 1).Int64()
}
// HIncrBy 为哈希表 key 中的指定字段的整数值加上增量 increment
func (rd *Redis) HIncrBy(key, field string, increment int64) int64 {
	return rd.pool.Do("HINCRBY "+key, field, increment).Int64()
}
// HDecrBy 为哈希表 key 中的指定字段的整数值减去增量 increment
func (rd *Redis) HDecrBy(key, field string, increment int64) int64 {
	return rd.pool.Do("HDECRBY "+key, field, increment).Int64()
}

// LPush 将一个或多个值插入到列表头部
// LPush return 成功添加的个数
func (rd *Redis) LPush(key string, values ...string) int {
	return rd.pool.Do("LPUSH", append(append([]interface{}{}, key), u.ToInterfaceArray(values)...)...).Int()
}
// RPush 在列表中添加一个或多个值
// RPush return 成功添加的个数
func (rd *Redis) RPush(key string, values ...string) int {
	return rd.pool.Do("RPUSH", append(append([]interface{}{}, key), u.ToInterfaceArray(values)...)...).Int()
}
// LPop 移出并获取列表的第一个元素
func (rd *Redis) LPop(key string) interface{} {
	return makeRedisResult(rd.pool.Do("LPOP " + key))
}
// RPop 移除并获取列表最后一个元素
func (rd *Redis) RPop(key string) interface{} {
	return makeRedisResult(rd.pool.Do("RPOP " + key))
}
// LLen 获取列表长度
// LLen 列表的长度
func (rd *Redis) LLen(key string) int {
	return rd.pool.Do("LLEN " + key).Int()
}
// LRange 获取列表指定范围内的元素
// LRange return []any 列表数据，如果值是一个对象则返回反序列化后的对象，否则返回字符串
func (rd *Redis) LRange(key string, start, stop int) []interface{} {
	return makeRedisResults(rd.pool.Do("LRANGE "+key, start, stop))
}

// TODO 支持订阅，还需要支持 Start、Stop、向Context注册析构函数确保Stop被执行，需要支持传入Function转化为func
// Subscribe
//func (rd *Redis) Subscribe(channel string, onReceived func([]byte)) bool {
//	return rd.pool.Subscribe(channel, nil, onReceived)
//}
//
// Unsubscribe
//func (rd *Redis) Unsubscribe(channel string) bool {
//	return rd.pool.Unsubscribe(channel)
//}

// Publish 将信息发送到指定的频道
// Publish channel 渠道名称
// Publish data 数据，字符串格式
func (rd *Redis) Publish(channel, data string) bool {
	return rd.pool.Do("PUBLISH "+channel, data).Bool()
}
