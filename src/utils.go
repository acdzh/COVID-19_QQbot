package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"main/cqp"
)

/*********************************支持函数 start****************************************/
func sendMsg(msg string, strategy int) { // 发送消息
	if globalRunMode == runModeDevInConsole {
		return
	}
	if globalRunMode == runModeDevOnLocalMachine {
		strategy = onlySendToPrivateDevStrategy
	}

	groupStrategy := strategy / 10
	privateStrategy := strategy % 10

	if groupStrategy == sendTOUserOnly || groupStrategy == sendToUserAndDev {
		for _, groupID := range userQQGroupIDs {
			cqp.SendGroupMsg(groupID, msg)
		}
	}
	if groupStrategy == sendToDevOnly || groupStrategy == sendToUserAndDev {
		for _, groupID := range devQQGroupIDs {
			cqp.SendGroupMsg(groupID, msg)
		}
	}
	if privateStrategy == sendTOUserOnly || privateStrategy == sendToUserAndDev {
		for _, qqID := range userQQIds {
			cqp.SendPrivateMsg(qqID, msg)
		}
	}
	if privateStrategy == sendToDevOnly || privateStrategy == sendToUserAndDev {
		for _, qqID := range devQQIds {
			cqp.SendPrivateMsg(qqID, msg)
		}
	}
}

func peopleCountsToString(counts map[string]string) string {
	s := ""
	for _, field := range countsField {
		s += fmt.Sprintf("%s / ", counts[field])
	}
	return s[:len(s)-2]
}

func timeStampToString(t int64) string { // 时间格式化
	return time.Unix(t/1000, 0).Format("2006-01-02 15:04:05 (北京时间)")
}

func itos(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func checkTimeInterval(t1, t2 float64) bool { // 检查时间间隔是否超过一小时
	var forCheckInterval float64 = shouldSendAllAfterUpgradeInterval * (60 * 1000) // ms
	defer writeLog(fmt.Sprintf("[checkTimeInterval] t1: %v, t2: %v, lastSendAllAfterUpgradeTimeStr: %v", t1, t2, lastSendAllAfterUpgradeTime))
	if t1-t2 >= forCheckInterval {
		lastSendAllAfterUpgradeTime = t1
		return true
	}
	if t2-t1 >= forCheckInterval {
		lastSendAllAfterUpgradeTime = t2
		return true
	}
	return false
}

func upgradeFormat(a, b string) string { // 格式化更新时的字符串
	writeLog(fmt.Sprintf("[upgradeFormat] a: \"%s\", b: \"%s\"", a, b))
	if a == b {
		writeLog(fmt.Sprintf("[upgradeFormat] return: \"%s\"", a))
		return a
	}
	if a == "" {
		a = "(新增)"
	} else if b == "" {
		b = "(删除)"
	}
	s := fmt.Sprintf("%s → %s", a, b)
	writeLog(fmt.Sprintf("[upgradeFormat] return: \"%s\"", s))
	return s
}

func isFileExisted(filename string) bool { // 文件是否存在
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func writeLog(l string) { // 写 log
	var filename string
	if globalRunMode == runModeDevInConsole {
		filename = "../log/" + time.Now().Format("2006-01-02") + ".log"
	} else {
		filename = logFilePath + time.Now().Format("2006-01-02") + ".log"
	}

	s := fmt.Sprintf("%v %s\n", time.Now().Format("15:04:05"), strings.Replace(l, "\n", "\"\n\"", -1))

	var f *os.File
	if isFileExisted(filename) {
		f, _ = os.OpenFile(filename, os.O_APPEND, 0666)
	} else {
		f, _ = os.Create(filename)
	}
	io.WriteString(f, s)

	if globalRunMode == runModeDevInConsole {
		fmt.Println(l)
	}

	f.Close()
}

func checkVer() { // 检查版本, 发更新日志
	var oldVersion string
	if isFileExisted(versionFileName) {
		content, _ := ioutil.ReadFile(versionFileName)
		oldVersion = string(content)
	} else {
		os.Create(versionFileName)
		oldVersion = "v0.0.0.0"
	}
	writeLog("[checkVer] current: " + currentVersion + ", old: " + oldVersion)

	if currentVersion != oldVersion {
		msgR := fmt.Sprintf("已更新: %s → %s\n\n更新日志: %s", oldVersion, currentVersion, versionUpgradeLog)
		writeLog("[checkVer] sendVersionUpgradeLogMsg: " + msgR)
		if shouldPushLog {
			sendMsg(msgR, versionSendStrategy)
		}
		f, _ := os.OpenFile(versionFileName, os.O_WRONLY|os.O_TRUNC, 0666)
		io.WriteString(f, currentVersion)
		f.Close()
	}
}
