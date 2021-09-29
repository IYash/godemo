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
	word1 := "hEllo"
	word2 := "中秋He"
	fmt.Println(strings.ToLower(word1))
	fmt.Println(word1)
	fmt.Println(strings.ToLower(word2))
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
		UserId:      "5fdc6cb90000000000002258",
		Content:     "中秋快乐",
		PostMilSec:  "4000",
		CommonType:  "0",
	}
	response := biz.CommonActivityStyleResolve(&option)
	fmt.Println(response.ActivityStyle)
	fmt.Println(response.StyleName)
}
