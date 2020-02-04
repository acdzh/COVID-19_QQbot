package main

import (
	"fmt"
	"strings"
)

/*********************************数据类型 dxyDatas start****************************************/

type dxyDatas map[string]interface{}

func itos(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func (d dxyDatas) toString() string {
	s := ""
	for _, arr := range neededAttributes {
		lineHead := arrHead[arr]
		lineBody := itos(d[arr])
		if lineHead != "" || lineBody != "" {
			s += (lineHead + lineBody + "\n")
		}
	}
	writeLog("[toString] s: " + strings.Replace(s, "\n", "\\n", 0))
	return strings.TrimRight(s, "\n")
}

func (d dxyDatas) toStringBeforeUpgrade(new dxyDatas) string {
	shouldShowAll := checkTimeInterval(int(new["modifyTime"].(float64)), lastSendAllAfterUpgradeTime)
	writeLog(fmt.Sprintf("[toStringBeforeUpgrade] shouldShowAll: %v", shouldShowAll))
	s := ""
	for _, arr := range neededAttributes {
		lineHead := arrHead[arr]
		lineBody := upgradeFormat(itos(d[arr]), itos(new[arr]))
		if (lineHead != "" || lineBody != "") && (shouldShowAll || d[arr] != new[arr]) {
			s += (lineHead + lineBody + "\n")
		}
	}
	writeLog("[toStringBeforeUpgrade] s: " + strings.Replace(s, "\n", "\\n", 0))
	return strings.TrimRight(s, "\n")
}

func (d dxyDatas) shouldUpgrade(new dxyDatas) bool {
	for _, arr := range forCheckAttributes {
		if d[arr] != new[arr] {
			writeLog("[shouldUpgrade] true")
			return true
		}
	}
	writeLog("[shouldUpgrade] false")
	return false
}

func (d dxyDatas) upgrade(new dxyDatas) { // 更新数据
	for arr := range new {
		d[arr] = new[arr]
	}
	writeLog("[upgrade]")
}

func (d dxyDatas) dataFmt() { // 获取到初始数据后, 再进行一些加工
	d["createTimeStr"] = timeStampToString(int64(d["createTime"].(float64)))
	d["modifyTimeStr"] = timeStampToString(int64(d["modifyTime"].(float64)))
	d["dxyUrl"] = dxyURL
	d["tencentUrl"] = tencentURL
	d["version"] = currentVersion

	for _, t := range [...]string{"confirmed", "suspected", "serious", "dead", "cured"} {
		if d[t+"Incr"] == nil {
			d[t+"Count"] = fmt.Sprintf("%v (较昨日无变化)", d[t+"Count"])
		} else {
			if strings.Contains(itos(d[t+"Incr"]), "-") {
				d[t+"Count"] = fmt.Sprintf("%v (较昨日 %v)", d[t+"Count"], d[t+"Incr"])
			} else {
				d[t+"Count"] = fmt.Sprintf("%v (较昨日 +%v)", d[t+"Count"], d[t+"Incr"])
			}
		}
	}

}
