package main

import (
	"fmt"
	"strings"
	"time"

	"main/cqp"
)

func main() {
	if globalRunMode == runModeDevInConsole {
		consoleTest()
	} else {
		cqp.Main()
	}
}

func init() {
	cqp.AppID = appid
	cqp.PrivateMsg = onPrivateMsg
	cqp.GroupMsg = onGroupMsg
	cqp.Enable = onEnable
	cqp.Exit = onExit
}

func onEnable() int32 {
	sendMsg("I am online!", onlineMsgSendStrategy)
	writeLog(fmt.Sprintf("%s", cqp.AppID))
	checkVer()
	d := prase(fetch())
	lastSendAllAfterUpgradeTime = d["modifyTime"].(float64)
	if lastNewsTimeStamp == 0 {
		lastNewsTimeStamp = time.Now().Unix()
	}
	writeLog(fmt.Sprintf("[onEnable] 初始化, lastSendAllAfterUpgradeTimeStr: %v", lastSendAllAfterUpgradeTime))
	go func(d dxyDatas) {
		for {
			if willPraseSuccess {
				writeLog("[onEnable] start check upgrade")
				current := prase(fetch())
				if d.shouldUpgrade(current) {
					msgR := d.toStringBeforeUpgrade(current)
					sendMsg(msgR, upgradeSendStrategy)
					d.upgrade(current)
				}
			}
			time.Sleep(refershInterval * time.Second)
		}
	}(d)
	go func() {
		for {
			writeLog(("[onEnable] start fetch news"))
			isUpdated, messages, err := parseNews(fetchNews())
			if err != nil {
				writeLog("[fetch news error]")
				continue
			}
			if !isUpdated {
				writeLog(("[onEnable] news ! up"))
				continue
			}
			for _, message := range messages {
				sendMsg(message, newsUpgradeSendStrategy)
				time.Sleep(2 * time.Second)
			}
			time.Sleep(newsRefershInterval * time.Second)
		}
	}()
	return 0
}

// 私聊发送任何消息都会回复当前情况
func onPrivateMsg(subType, msgID int32, fromQQ int64, msg string, font int32) int32 {
	writeLog(fmt.Sprintf("[onPrivateMsg] %d %s", fromQQ, msg))
	msgR := whatToReply(msg)
	cqp.SendPrivateMsg(fromQQ, msgR)
	return 0
}

// 群聊中@bot即回复当前情况
func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if strings.Contains(msg, "[CQ:at,qq="+selfQQID+"]") {
		writeLog(fmt.Sprintf("[onGroupMsg] %d %d %s", fromGroup, fromQQ, msg))
		msgR := whatToReply(strings.Replace(msg, "[CQ:at,qq="+selfQQID+"]", "", -1))
		cqp.SendGroupMsg(fromGroup, msgR)
	}
	return 0
}

func whatToReply(msg string) string {
	msg = strings.TrimLeft(msg, " ")
	if len(msg) > 2 && msg[0] == 'q' && msg[1] == ' ' {
		global, area, err := htmlGetAllProvinceAndCity(fetch())
		if err != nil {
			return fmt.Sprintf("%v", err)
		}
		replyMsg := ""
		msg = strings.TrimLeft(msg[1:], " ")
		for _, cityOrProvinceName := range strings.Split(msg, " ") {
			if cityOrProvinceName != " " && cityOrProvinceName != "" {
				d, name, err := getDatasOfProvinceOrCity(global, area, cityOrProvinceName)
				if err != nil {
					replyMsg += fmt.Sprintf("%v\n", err)
				} else {
					replyMsg += fmt.Sprintf("%s: %s\n", name, peopleCountsToString(d))
				}
			}
		}
		return strings.TrimRight(replyMsg, "\n")
	}
	if len(msg) > 3 && msg[0] == 'q' && msg[1] == 'a' && msg[2] == ' ' {
		chinaAllDatas, err := htmlGetAreaStat(fetch())
		if err != nil {
			return fmt.Sprintf("%v", err)
		}
		replyMsg := ""
		msg = strings.TrimLeft(msg[2:], " ")
		for _, cityOrProvinceName := range strings.Split(msg, " ") {
			if cityOrProvinceName != " " && cityOrProvinceName != "" {
				t, err := getDataStrsOfCitesOfAProvince(chinaAllDatas, cityOrProvinceName)
				if err != nil {
					replyMsg += fmt.Sprintf("%v\n\n", err)
				} else {
					replyMsg += fmt.Sprintf("%s\n\n", t)
				}
			}
		}
		return strings.TrimRight(replyMsg, "\n")
	}
	if len(msg) > 3 && msg[0] == 'd' && strings.Contains(msg, "debug") {
		s := ""
		for k, v := range prase(fetch()) {
			s += fmt.Sprintf("%v: %v\n", k, v)
		}
		return s
	}
	if len(msg) > 2 && msg[0] == 'u' && strings.Contains(msg, "url") {
		return urlList
	}
	return prase(fetch()).toString()
}

func onExit() int32 {
	writeLog("exit !")
	sendMsg("exit !", failedDataSendStrategy)
	return 0
}
