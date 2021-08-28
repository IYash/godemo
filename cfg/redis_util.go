package cfg

import (
	"github.com/go-redis/redis/v7"
	"github.com/wonderivan/logger"
	"time"
)
var client *redis.Client
func InitClient(){
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379",Password: "",DB: 0})
	pong,_ := client.Ping().Result()
	logger.Info("Redis ping:", pong)
}
func Set(key string,value string) {
	client.Set(key,value,time.Duration(180)*time.Second)
}
func HMset(key string,field string,value string) {
	client.HMSet(key,field,value)
}
func Hget(key string,field string) string {
	fieldVal,_ := client.HGet(key,field).Result()
	return fieldVal
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
