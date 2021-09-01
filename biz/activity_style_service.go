package biz
import (
	"encoding/json"
	"godemo/cfg"
	"godemo/serialize"
	"strconv"
	"time"
)

const (

	//通用弹幕样式
	CommonActivityName = "danmaku_activity_names"  //[活动1，活动2]
	CommonActivityRuleInfo = "hash:danmaku_activity_rule_info" //遍历ActivityName来完成，实现同时兼容某种形式的活动样式
	CommonActivityTopics = "danmaku_activity_topics"
	CommonActivityNotes = "danmaku_activity_notes"
	CommonActivityCondition = "danmaku_activity_condition" //0:none,1:topic,2:note 组合数值，通过时刻区分是否是官方弹幕
	CommonActivityType = "danmaku_activity_type"           //引导类型：1，通用类型：0
	CommonActivityUserId = "user_id"
	CommonActivityPostTime = "post_milsec"
	CommonActivityBeginTime = "danmaku_activity_begin_time"
	CommonActivityEndTime = "danmaku_activity_end_time"
	CommonActivitySwitch = "danmaku_activity_switch"
	CommonActivityEffectV = "danmaku_activity_effect_v"
	CommonActivitySpecialContent = "danmaku_activity_special_content"
	CommonActivityStyleInfo = "hash:danmaku_activity_style_info"
)

type DanmakuActivityStyleOption struct {

	NoteId      string
	Content     string
	TopicIds    []string
}

type DanmakuOperatorStyleOption struct {

	UserId      string
	Content     string
	PostMilSec  string
}

type CommonActivityStyleOption struct {

	UserId      string
	Content     string
	PostMilSec  string
	NoteId      string
	TopicIds    []string
	CommonType	string           //通用0,引导1
}


func activityDurationCheck(hotKeyFromRedis map[string]string,beginTimeKey string,endTimeKey string) bool {
	beginTime := hotKeyFromRedis[beginTimeKey]
	endTime := hotKeyFromRedis[endTimeKey]
	begin, e1 := strconv.ParseInt(beginTime, 10, 64)
	if e1 != nil {
		return false
	}
	end, e2 := strconv.ParseInt(endTime, 10, 64)
	if e2 != nil {
		return false
	}
	now := time.Now().Unix()
	return now >= begin && now <= end
}

func topicMatchCheck(hotKeyFromRedis map[string]string, option *CommonActivityStyleOption) bool {
	//指定话题判断
	topics := []byte(hotKeyFromRedis[CommonActivityTopics])
	var topicRaw []string
	json.Unmarshal(topics, &topicRaw)
	topicIds := option.TopicIds
	return  arrayMatch(topicRaw, topicIds)
}
func noteMatchCheck(hotKeyFromRedis map[string]string, option *CommonActivityStyleOption) bool {
	//指定笔记判断
	noteIds := []byte(hotKeyFromRedis[CommonActivityNotes])
	var noteRaw []string
	json.Unmarshal(noteIds, &noteRaw)
	noteids := []string{}
	noteids = append(noteids, option.NoteId)
	return arrayMatch(noteRaw, noteids)
}
func arrayMatch(source []string, sub []string) bool {
	sourceLen := len(source)
	subLen := len(sub)
	if sourceLen == 0 || subLen == 0 {
		return false
	}
	//基于source,sub数量极其有限
	for i := 0; i < sourceLen; i++ {
		for j := 0; j < subLen; j++ {
			if source[i] == sub[j] {
				return true
			}
		}
	}
	return false
}


func fillCommonActivityStyle(option *CommonActivityStyleOption,activityName string) string {
	styleKey := activityName + CommonActivityStyleInfo
	hotKeyFromRedis,_ := cfg.HgetAll(styleKey)
	return hotKeyFromRedis[option.Content]
}
func isStyleAllow(hotKeyFromRedis map[string]string, ruleSwitch string, effectV string, content string) bool {
	if nil == hotKeyFromRedis {
		return false
	}
	if len(content) == 0 {
		return false
	}
	//活动开关，默认开启
	if "0" == hotKeyFromRedis[ruleSwitch] {
		return false
	}

	return true
}


func CommonSpecialContentForActivity() string {
	var replaceContent string
	activityNames := cfg.Get(CommonActivityName)
	if serialize.StringIsEmpty(activityNames) {
		return ""
	}
	var activityNameArr []string
	json.Unmarshal([]byte(activityNames),&activityNameArr)
	for _,activityName := range activityNameArr {
		activityKey := activityName + CommonActivityRuleInfo
		hotKeyFromRedis,_ := cfg.HgetAll(activityKey)
		//先判断样式是作用于哪类活动的
		activityType := hotKeyFromRedis[CommonActivityType]
		if activityType == "0" {
			replaceContent = hotKeyFromRedis[CommonActivitySpecialContent]
			if !serialize.StringIsEmpty(replaceContent) {
				break
			}
		}
	}
	return replaceContent
}
//弹幕样式消费，以活动为基本单位，支持到用户/话题；用户/笔记维度，调用方可能需要调用2次
func CommonActivityStyleResolve(option *CommonActivityStyleOption) (*string,int) {
	activityStyle := ""
	styleFrom := -1
	activityNames := cfg.Get(CommonActivityName)
	if serialize.StringIsEmpty(activityNames) {
		return &activityStyle,styleFrom
	}
	var activityNameArr []string
	json.Unmarshal([]byte(activityNames),&activityNameArr)
	for _,activityName := range activityNameArr {
		activityKey := activityName + CommonActivityRuleInfo
		hotKeyFromRedis,_ := cfg.HgetAll(activityKey)
		//先判断样式是作用于哪类活动的
		activityType := hotKeyFromRedis[CommonActivityType]
		if option.CommonType != activityType {
			continue
		}
		if !isStyleAllow(hotKeyFromRedis, CommonActivitySwitch, CommonActivityEffectV, option.Content) {
			return &activityStyle,styleFrom
		}
		if !activityDurationCheck(hotKeyFromRedis,CommonActivityBeginTime,CommonActivityEndTime) {
			return &activityStyle,styleFrom
		}
		var match bool
		if option.CommonType == "0" { //通用样式
			match,_ = conditionMatchAny(hotKeyFromRedis,option)
			if match {
				activityStyle = fillCommonActivityStyle(option,activityName)
			}
		} else { //引导样式
			match,styleFrom = conditionMatchScope(hotKeyFromRedis,option)
			if match {
				activityStyle = fillCommonActivityStyle(option,activityName)
			}
		}
		if !serialize.StringIsEmpty(activityStyle) {
			break
		}
	}
	return &activityStyle,styleFrom
}
func isGuideDanmakuCheck(hotKeyFromRedis map[string]string,option *CommonActivityStyleOption) bool{
	userId := option.UserId
	postMilSec := option.PostMilSec
	return userId == hotKeyFromRedis[CommonActivityUserId] && postMilSec == hotKeyFromRedis[CommonActivityPostTime]
}
//全量样式配置时的条件检查
func conditionMatchAny(hotKeyFromRedis map[string]string,option *CommonActivityStyleOption) (bool,int) {
	conditionVal := hotKeyFromRedis[CommonActivityCondition]
	if !serialize.StringIsEmpty(conditionVal) {
		val,_ := strconv.Atoi(conditionVal)
		if val == 0 {  //没有任何条件限制,适用于全量样式的场景
			return true,-1
		}
		if val & 1 == 1  { //笔记限制或者话题限制
			if topicMatchCheck(hotKeyFromRedis,option) {
				return true,1
			}
		}
		if val & 2 == 2 {
			if noteMatchCheck(hotKeyFromRedis,option){
				return true,2
			}
		}
	}
	return false,-1
}
//引导样式配置时的条件检查
func conditionMatchScope(hotKeyFromRedis map[string]string,option *CommonActivityStyleOption) (bool,int) {
	if isGuideDanmakuCheck(hotKeyFromRedis,option) {
		return conditionMatchAny(hotKeyFromRedis,option)
	}
	return false,-1
}
