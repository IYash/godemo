package cfg

import (
	"github.com/go-redis/redis/v7"
	"github.com/wonderivan/logger"
	"time"
)
var client *redis.Client
func init(){
	InitClient()
}
func InitClient(){
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379",Password: "",DB: 0})
	pong,_ := client.Ping().Result()
	logger.Info("Redis ping:", pong)
}
func Set(key string,value interface{}) {
	client.Set(key,value,time.Duration(180)*time.Second)
}
func Get(key string) string {
	val,_ := client.Get(key).Result()
	return val
}
func HSet(key string,field string,value interface{}) {
	client.HMSet(key,field,value)
}
func HGet(key string,field string) string {
	fieldVal,_ := client.HGet(key,field).Result()
	return fieldVal
}
func HgetAll(key string) (map[string]string,error){
	return client.HGetAll(key).Result()
}
func GetAndDel(key string) string {
	luastr := "local current = redis.call('get',KEYS[1]); if(current) then redis.call('del',KEYS[1]); end return current;"
	script :=redis.NewScript(luastr)
	n,err :=script.Run(client,[]string{key}).Result()
	if err !=nil{
		logger.Warn("lua eval error:",err)
		return ""
	}
	if n == nil {
		return ""
	}
	return n.(string)
}
func SIsMember(key string,data string) bool {
	isMember,_ := client.SIsMember(key,data).Result()
	return isMember
}
func SAdd(key string,data string) {
	client.SAdd(key,data)
}
func Decr(key string) {
	client.Decr(key)
}
