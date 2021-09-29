package biz
import (
	"encoding/json"
	"fmt"
	"godemo/cfg"
	"godemo/serialize"
	"strconv"
	"strings"
	"time"
)

const (

	CommonActivityName           = "danmaku_activity_names"          //[活动1，活动2]
	CommonActivityRuleInfo       = "hash:danmaku_activity_rule_info" //遍历ActivityName来完成，实现同时兼容某种形式的活动样式
	CommonActivityTopics         = "danmaku_activity_topics"
	CommonActivityNotes          = "danmaku_activity_notes"
	CommonActivityCondition      = "danmaku_activity_condition" //0:none,1:topic,2:note 组合数值，通过时刻区分是否是官方弹幕
	CommonActivityType           = "danmaku_activity_type"      //引导类型：1，通用类型：0
	CommonActivityUserId         = "user_id"
	CommonActivityPostTime       = "post_milsec"
	CommonActivityBeginTime      = "danmaku_activity_begin_time"
	CommonActivityEndTime        = "danmaku_activity_end_time"
	CommonActivitySwitch         = "danmaku_activity_switch"
	CommonActivityEffectV        = "danmaku_activity_effect_v"
	CommonActivitySpecialContent = "danmaku_activity_special_content"
	CommonActivityStyleInfo      = "hash:danmaku_activity_style_info"

	CommonActivityMulStyleInfo = "hash:danmaku_activity_mul_style_info"
)

type CommonActivityStyleOption struct {

	UserId      string
	Content     string
	PostMilSec  string
	NoteId      string
	TopicIds    []string
	CommonType  string //通用0,引导1
	NoteUserId  string //用于作者弹幕样式化
	StyleName   string
}
type CommonActivityStyleResponse struct {
	ActivityStyle string
	StyleFrom     int
	StyleName     string
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
func CommonActivityStyleResolve(option *CommonActivityStyleOption) *CommonActivityStyleResponse {
	commonActivityStyleResponse := &CommonActivityStyleResponse{
		ActivityStyle: "",
		StyleFrom:     -1,
		StyleName:     "",
	}
	activityStyle := ""
	styleFrom := -1
	styleName := ""
	activityNames := cfg.Get(CommonActivityName)
	if serialize.StringIsEmpty(activityNames) {
		return commonActivityStyleResponse
	}
	var activityNameArr []string
	json.Unmarshal([]byte(activityNames), &activityNameArr)
	for _, activityName := range activityNameArr {
		activityKey := activityName + CommonActivityRuleInfo
		hotKeyFromRedis ,_:= cfg.HgetAll(activityKey)
		//先判断样式是作用于哪类活动的
		activityType := hotKeyFromRedis[CommonActivityType]
		if option.CommonType != activityType {
			continue
		}
		if !isStyleAllow(hotKeyFromRedis, CommonActivitySwitch, CommonActivityEffectV, option.Content) {
			continue
		}
		if !activityDurationCheck(hotKeyFromRedis, CommonActivityBeginTime, CommonActivityEndTime) {
			continue
		}
		var match bool
		if option.CommonType == "0" { //通用样式
			match, _ = conditionMatchAny(hotKeyFromRedis, option)
			if match {
				activityStyle = fillCommonActivityStyle(option, activityName)
			}
		} else { //引导样式
			match, styleFrom = conditionMatchScope(hotKeyFromRedis, option)
			if match {
				activityStyle = fillCommonActivityStyle(option, activityName)
			}
		}
		if serialize.StringIsNotEmpty(activityStyle) { //这里有可能是styleName
			if !strings.Contains(activityStyle, ",") {
				if specialStyleTrigger(option.UserId) {  //
					styleName = "midAutumn-special"
				} else {
					styleName = "midAutumn-common"
				}
				activityStyle = DanmakuActivityStyleFromStyleName(&CommonActivityStyleOption{CommonType:"0",StyleName: styleName})
			}
			//如果是中秋样式，需要返回styleName
			commonActivityStyleResponse.ActivityStyle = activityStyle
			commonActivityStyleResponse.StyleFrom = styleFrom
			commonActivityStyleResponse.StyleName = styleName
			break
		}
	}
	return commonActivityStyleResponse
}
func isGuideDanmakuCheck(hotKeyFromRedis map[string]string,option *CommonActivityStyleOption) bool{
	userId := option.UserId
	postMilSec := option.PostMilSec
	noteUserId := option.NoteUserId
	return (userId == hotKeyFromRedis[CommonActivityUserId] && postMilSec == hotKeyFromRedis[CommonActivityPostTime]) || userId == noteUserId

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

func DanmakuActivityStyleFromStyleName(option *CommonActivityStyleOption) string { //只处理common
	activityStyle := ""
	activityNames := cfg.Get(CommonActivityName)
	if serialize.StringIsEmpty(activityNames) {
		return activityStyle
	}
	var activityNameArr []string
	json.Unmarshal([]byte(activityNames), &activityNameArr)
	for _, activityName := range activityNameArr {
		activityKey := activityName + CommonActivityRuleInfo
		hotKeyFromRedis,_ := cfg.HgetAll( activityKey)
		if nil == hotKeyFromRedis {
			continue
		}
		activityType := hotKeyFromRedis[CommonActivityType]
		if activityType != option.CommonType {
			continue
		}
		if !activityDurationCheck(hotKeyFromRedis, CommonActivityBeginTime, CommonActivityEndTime) {
			continue
		}
		mulStyleKey := activityName + CommonActivityMulStyleInfo
		mulStyleFromRedis,_ := cfg.HgetAll(mulStyleKey)
		if nil == mulStyleFromRedis {
			continue
		}
		activityStyle = mulStyleFromRedis[option.StyleName]
		if serialize.StringIsNotEmpty(activityStyle) {
			break
		}
	}
	return activityStyle
}
//中秋官方弹幕管理，只有活动期内可见
func DanmakuNeedHide(option *CommonActivityStyleOption) bool {

	activityNames := cfg.Get( CommonActivityName)
	if serialize.StringIsEmpty(activityNames) {
		return false
	}
	var activityNameArr []string
	json.Unmarshal([]byte(activityNames), &activityNameArr)
	for _, activityName := range activityNameArr {
		if strings.Contains(activityName, "midAutumn") {
			activityKey := activityName + CommonActivityRuleInfo
			hotKeyFromRedis,_ := cfg.HgetAll(activityKey)
			//先判断样式是作用于哪类活动的
			if nil == hotKeyFromRedis {
				return false
			}
			if !activityDurationCheck(hotKeyFromRedis, CommonActivityBeginTime, CommonActivityEndTime) {
				guideUserId := hotKeyFromRedis[CommonActivityUserId] //暂时只通过账号来限定
				if guideUserId == option.UserId {
					return true
				}
			}
		}
	}
	return false
}

//活动话题下的笔记预置引导弹幕,会成为视频的第一条弹幕,注意并发问题
func DanmakuAutoGuide(option *CommonActivityStyleOption) {

	activityNames := cfg.Get(CommonActivityName)
	if serialize.StringIsEmpty(activityNames) {
		return
	}
	var activityNameArr []string
	json.Unmarshal([]byte(activityNames), &activityNameArr)
	for _, activityName := range activityNameArr {
		if strings.Contains(activityName, "midAutumn") {
			activityKey := activityName + CommonActivityRuleInfo
			hotKeyFromRedis,_ := cfg.HgetAll(activityKey)
			if nil == hotKeyFromRedis {
				continue
			}
			activityType := hotKeyFromRedis[CommonActivityType]
			if !activityDurationCheck(hotKeyFromRedis, CommonActivityBeginTime, CommonActivityEndTime) {
				continue
			}
			if activityType == "1" { //引导类型
				//话题匹配
				if !topicMatchCheck(hotKeyFromRedis, option) {
					continue
				}
				styleKey := activityName + CommonActivityStyleInfo
				styleInfo,_ := cfg.HgetAll(styleKey)
				if nil == styleInfo {
					continue
				}
				for k := range styleInfo {
					//构建引导弹幕
					fmt.Println(k)
				}
			}
		}
	}
}
func specialStyleTrigger(userId string) bool { //如上是满足了样式的条件，多样式下判断是否触发特殊样式
	timeObj := time.Now()
	year := timeObj.Year()
	month := timeObj.Month()
	day := timeObj.Day()
	dateInfo := fmt.Sprintf("%d-%d-%d", year, month, day)
	//先判断当前用户在当日是否已经触发过

	userKey := fmt.Sprintf(cfg.MidAutumnTriggerUserKey, dateInfo)
	exist := cfg.SIsMember(userKey, userId)
		if exist {
			return false
		}
	currentTime := timeObj.Unix()
	remainKey := fmt.Sprintf(cfg.MidAutumnDayUsedKey, dateInfo)
	remainVal := remainAccumulate(year, month, day)
	if remainVal > 0 {
		lastTime  := cfg.HGet(cfg.MidAutumnTriggerLastKey,dateInfo)
		if serialize.StringIsEmpty(lastTime) {
			return false
		}
		lastTimeVal, e3 := strconv.Atoi(lastTime)
		pastTime := currentTime - int64(lastTimeVal)
		if e3 == nil  && pastTime > 0 && (currentTime%remainVal == 0 || pastTime > 144/remainVal+1) {
			cfg.HSet(cfg.MidAutumnTriggerLastKey, dateInfo,currentTime)
			cfg.Decr(remainKey)
			cfg.SAdd(userKey, userId)
			return true
		}
	}
	return false
}
func remainAccumulate(year int, month time.Month, day int) int64 {
	remainVal := 0
	curVal := 0
	curDay := day
	round := 3
	for round > 0 {
		round = round - 1
		curVal = remainCalculate(year, month, curDay)
		remainVal = remainVal + curVal
		curDay = curDay - 1
	}
	return int64(remainVal)
}

func remainCalculate(year int, month time.Month, day int) int { //获取每天的remain
	remainVal := 0
	dateInfo := fmt.Sprintf("%d-%d-%d", year, month, day)
	remainKey := fmt.Sprintf(cfg.MidAutumnDayUsedKey, dateInfo)
	remain:= cfg.Get(remainKey)
	remainV, err2 := strconv.Atoi(remain)
	if err2 == nil {
		remainVal = remainV
	}
	return remainVal
}