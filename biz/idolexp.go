package biz

import (
	"encoding/json"
	"fmt"
	"godemo/cfg"
	"godemo/serialize"
	"strings"
)

func MatchIdolExp(noteId string,userId string,content string,exp int32) bool {
	//获取idol列表
	if !DanmakuContentMatch(content) { //非idol相关内容
		return true
	}
	//明星弹幕分发实验
	idolContentExp := exp

	if idolContentExp == 0 {
		return true
	}
	if idolContentExp ==1 {
		return false
	}
	//获取用户及笔记的特征信息
	var profileInfo map[string]string
	var userAge int32
	//var wg1 sync.WaitGroup
	//wg1.Add(1)
	//go func() {
	//	defer wg1.Done()
	//	//用户特征
	//	var userIds []string
	//	userIds = append(userIds,userId)
	//	userAge = BatchGetUserAges()
	//}()
	//go func() {
	//	defer wg1.Done()
	//	//笔记特征
	//	profileInfo = GetNoteProfileTag(noteId)
	//}()
	//wg1.Wait()
	userAge = 10
	profileInfo = GetNoteProfileTag(noteId)
	if idolContentExp ==2 && userAge<=18 && userAge >0 {
		return true
	}
	if idolContentExp ==3 && userAge<=18 && userAge >0 {
		categoryIds := [3]string{"c00000000000000000000394","5ab1c90e3a03eb81765398d5","c00000000000000000000376"}
		for _,val := range categoryIds {
			if val == profileInfo["np_tax2"] {
				return true
			}
		}
		return false
	}
	return true
}
var IdolInfoForDanmakuKey = "string:danmaku_idol_info"
var NoteTaxonomyKey ="string:note-service:note_taxonomy:%s"
func DanmakuContentMatch( content string) bool{
	idolListInfo := cfg.Get(IdolInfoForDanmakuKey)
	if serialize.StringIsEmpty(idolListInfo) {
		return false
	}
	var contentRaw []string
	json.Unmarshal([]byte(idolListInfo), &contentRaw)
	for _,v := range contentRaw {
		if strings.Contains(content,v) {
			return true
		}
	}
	return false
}
func BatchGetUserAges() int32{
	return 10
}

func GetNoteProfileTag(noteId string) map[string]string {
	result := map[string]string{}
	key := fmt.Sprintf(NoteTaxonomyKey, noteId)
	cacheTaxonomy := cfg.Get(key)

	if  !serialize.StringIsEmpty(cacheTaxonomy) {
		result["np_tax2"] = cacheTaxonomy
		return result
	}
	if serialize.StringIsEmpty(noteId) {
		return result
	}
	result["np_tax2"] = "c00000000000000000000394"
	cfg.Set(key,result["np_tax2"])

	return result
}