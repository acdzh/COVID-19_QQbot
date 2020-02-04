package main

import (
	"fmt"
	"strings"
	"time"

	"main/cqp"
)

func consoleTest() {
	html := fetch()
	d := prase(html)
	lastSendAllAfterUpgradeTime = d["modifyTime"].(float64)
	writeLog(fmt.Sprintf("%v\n", lastSendAllAfterUpgradeTime))
	dd := prase(html)
	dd["deadCount"] = "9843 (较昨日 +59)"
	dd["modifyTime"] = 1580722478000.0
	for k, v := range d {
		writeLog(fmt.Sprintf("%v, %v", k, v))
	}

	writeLog("")
	writeLog(d.toString())
	writeLog("")
	writeLog(d.toStringBeforeUpgrade(dd))
	d.upgrade(dd)
	writeLog("")
	writeLog(d.toString())
}

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
	return 0
}

// 私聊发送任何消息都会回复当前情况
func onPrivateMsg(subType, msgID int32, fromQQ int64, msg string, font int32) int32 {
	writeLog(fmt.Sprintf("[onPrivateMsg] %d %s", fromQQ, msg))
	if strings.Contains(msg, "url") {
		cqp.SendPrivateMsg(fromQQ, urlList)
		return 0
	}
	msgR := prase(fetch()).toString()
	cqp.SendPrivateMsg(fromQQ, msgR)
	return 0
}

// 群聊中@bot即回复当前情况
func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if strings.Contains(msg, "[CQ:at,qq="+selfQQID+"]") {
		writeLog(fmt.Sprintf("[onGroupMsg] %d %d %s", fromGroup, fromQQ, msg))
		if strings.Contains(msg, "url") {
			cqp.SendGroupMsg(fromGroup, urlList)
			return 0
		} else if strings.Contains(msg, "debug") {
			s := ""
			for k, v := range prase(fetch()) {
				s += fmt.Sprintf("%v: %v\n", k, v)
			}
			cqp.SendGroupMsg(fromGroup, s)
			return 0
		}
		msgR := prase(fetch()).toString()
		cqp.SendGroupMsg(fromGroup, msgR)
	}
	return 0
}

func onExit() int32 {
	writeLog("exit !")
	sendMsg("exit !", failedDataSendStrategy)
	return 0
}
