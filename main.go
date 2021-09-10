package main

import (
	"fmt"
	"godemo/biz"
	"godemo/cfg"
	"godemo/model"
	"godemo/serialize"
	"strings"
)

func main(){
	fmt.Println(biz.MatchIdolExp("1a","2","李寻欢",3))
}
func join(source string) string{
	parts := strings.Split(source,",")
	return strings.Join(parts[:2],",")
}
func luaScript() {
	cfg.InitClient()
	cfg.Set("hello","world")
	fmt.Println(cfg.GetAndDel("hello"))
	fmt.Println(cfg.GetAndDel("hello"))
}


func convert(){
	zoo := model.BuildZoo("json")
	data,err := serialize.Marshal(zoo.Animals)
	if err!= nil {
		fmt.Println(err)
	}
	fmt.Println(serialize.BytesToString(data))
	var animals []model.Animal
	serialize.Unmarshal(data,&animals)
	fmt.Println(animals)
}
func activityStyle(){
	option := biz.CommonActivityStyleOption{
		UserId:      "5fdc6cb90000000000002249",
		Content:     "test0901",
		PostMilSec:  "4000",
		CommonType:  "1",
	}
	activityStyle,styleFrom := biz.CommonActivityStyleResolve(&option)
	fmt.Println(*activityStyle)
	fmt.Println(styleFrom)
}
