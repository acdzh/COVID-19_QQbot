package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Tnze/CoolQ-Golang-SDK/v2/cqp"
)

/*****************************自定义数据请在此处修改**********************************/

// 基本信息
const appid string = "com.acdzh.dxy"

//调试模式
const isDevMode bool = false

// 主动刷新间隔
const refershInterval = 5 // 分钟

// 自定义查询子区域 (未对所有地市进行匹配, 如果失败请自行修改正则
const provinceName string = "山东"
const cityName string = "菏泽"

// bot版本信息
const currentVersion string = "v1.27.20.15 beta"                                                                            // 当前版本, 每次修改后会进行版本更新推送
const versionUpgradeLog string = "1. 大幅优化代码结构, 提取复用组件\n2. 开源, 并增强了自定制功能\n3. 优化不同订阅权限管理, 优化消息发送逻辑\n4. TODO: 私聊消息订阅, 个性化城市订阅" // 版本更新日志, 仅会推送一次
const versionFileName string = "conf/dxy.cfg"                                                                               // 存储版本号
const logFilePath string = "data/log/"                                                                                      // log文件目录 (log会以日期命名
const shouldPushLog bool = true                                                                                             // 是否在每次更新之后更新版本推送

// url
const dxyURL string = "https://3g.dxy.cn/newh5/view/pneumonia" // 数据来源url
const devURL string = "http://127.0.0.1:5500/index.html"       // 本地调试url

// qqGroup & qqID
var selfQQID string = "1472745738"                    // bot自己的qq号
var userQQGroupIDs = [...]int64{854378285, 361684286} // 普通用户qq群数组
var devQQGroupIDs = [...]int64{584405782}             // 开发者调试用qq群数组
var userQQIds = [...]int64{}                          // 普通用户订阅qq号数组
var devQQIds = []int64{1069436872}                    // 开发者qq号数组

// 消息发送策略模板, 不要修改
const sendToNobody int = 0     // 不发送给任何类型用户或群组
const sendToUserAndDev int = 1 // 同时发送给普通和管理员用户或群组
const sendTOUserOnly int = 2   // 仅发送给普通用户或群组
const sendToDevOnly int = 3    // 仅发送给管理员用户或群组

// 具体的消息发送策略 (格式为: 10 * 群消息策略 + 私聊消息策略
const onlineMsgSendStrategy int = 10*sendToNobody + sendToDevOnly      // 上线提醒: 仅私聊发给管理员账号
const firstDataSendStrategy int = 10*sendToDevOnly + sendToNobody      // 上线后拉取的初始数据: 仅发送到调试qq群
const versionSendStrategy int = 10*sendToUserAndDev + sendToDevOnly    // 版本日志: 发送给所有群, 但私聊仅发送给管理员
const upgradeSendStrategy int = 10*sendToUserAndDev + sendToUserAndDev // 数据更新: 发送给所有群和用户

/*****************************自定义数据请在此处修改**********************************/

func sendMsg(msg string, strategy int) {
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

type dxyDatas struct {
	ddlTime           string
	confirmedNumber   string
	suspectedNumber   string
	deadNumber        string
	cureNumber        string
	provinceNumber    string
	cityNumber        string
	susceptiblePeople string
	incubation        string
	spreadWay         string
	isChanged         string
	diffusion         string
}

func (d dxyDatas) toString() string {
	return fmt.Sprintf(
		"更新时间: %s\n确诊: %s\n疑似: %s\n死亡: %s\n治愈: %s\n%s省: %s\n%s市: %s\n传染源: %s\n病毒: %s\n传播途径: %s\n易感人群: %s\n潜伏期: %s\n数据来源: 丁香园\nbot当前版本: %s",
		d.ddlTime,
		d.confirmedNumber,
		d.suspectedNumber,
		d.deadNumber,
		d.cureNumber,
		provinceName,
		d.provinceNumber,
		cityName,
		d.cityNumber,
		d.susceptiblePeople,
		d.incubation,
		d.spreadWay,
		d.isChanged,
		d.diffusion,
		currentVersion)
}

func (d dxyDatas) toStringWithOutTime() string {
	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s", d.confirmedNumber, d.suspectedNumber, d.deadNumber, d.cureNumber, d.provinceNumber, d.cityNumber, d.susceptiblePeople, d.incubation, d.spreadWay, d.isChanged, d.diffusion)
}

func (d dxyDatas) toStringAfterUpgrade(new dxyDatas) string {
	upgradeFormat := func(a, b string) string {
		if a == b {
			return a
		}
		return fmt.Sprintf("%s -> %s (已更新)", a, b)
	}
	return fmt.Sprintf(
		"传染数据已更新!\n更新时间: %s\n确诊: %s\n疑似: %s\n死亡: %s\n治愈: %s\n%s省: %s\n%s市: %s\n传染源: %s\n病毒: %s\n传播途径: %s\n易感人群: %s\n潜伏期: %s\n数据来源: 丁香园\nbot当前版本: %s",
		upgradeFormat(d.ddlTime, new.ddlTime),
		upgradeFormat(d.confirmedNumber, new.confirmedNumber),
		upgradeFormat(d.suspectedNumber, new.suspectedNumber),
		upgradeFormat(d.deadNumber, new.deadNumber),
		upgradeFormat(d.cureNumber, new.cureNumber),
		provinceName,
		upgradeFormat(d.provinceNumber, new.provinceNumber),
		cityName,
		upgradeFormat(d.cityNumber, new.cityNumber),
		upgradeFormat(d.susceptiblePeople, new.susceptiblePeople),
		upgradeFormat(d.incubation, new.incubation),
		upgradeFormat(d.spreadWay, new.spreadWay),
		upgradeFormat(d.isChanged, new.isChanged),
		upgradeFormat(d.diffusion, new.diffusion),
		currentVersion)
}

func (d *dxyDatas) shouldUpgrade(new *dxyDatas) bool {
	return !(d.toStringWithOutTime() == new.toStringWithOutTime())
}

func (d *dxyDatas) upgrade(new *dxyDatas) {
	d.ddlTime = new.ddlTime
	d.confirmedNumber = new.confirmedNumber
	d.suspectedNumber = new.suspectedNumber
	d.deadNumber = new.deadNumber
	d.cureNumber = new.cureNumber
	d.susceptiblePeople = new.susceptiblePeople
	d.incubation = new.incubation
	d.spreadWay = new.spreadWay
	d.isChanged = new.isChanged
	d.diffusion = new.diffusion
	d.provinceNumber = new.provinceNumber
	d.cityNumber = new.cityNumber
}

func isFileExisted(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func writeLog(l string) {
	filename := logFilePath + time.Now().Format("2006-01-02") + ".log"
	s := fmt.Sprintf("%v %s\n", time.Now().Format("15:04:05"), strings.Replace(l, "\n", "\"\n\"", -1))

	var f *os.File
	if isFileExisted(filename) {
		f, _ = os.OpenFile(filename, os.O_APPEND, 0666)
	} else {
		f, _ = os.Create(filename)
	}
	io.WriteString(f, s)

	f.Close()
}

func checkVer() {
	var oldVersion string
	if isFileExisted(versionFileName) {
		content, _ := ioutil.ReadFile(versionFileName)
		oldVersion = string(content)
	} else {
		os.Create(versionFileName)
		oldVersion = "v0.0.0.0"
	}

	if currentVersion != oldVersion {
		msgR := fmt.Sprintf("bot已更新: %s -> %s\n\n更新日志: %s", oldVersion, currentVersion, versionUpgradeLog)
		sendMsg(msgR, versionSendStrategy)
		f, _ := os.OpenFile(versionFileName, os.O_WRONLY, 0666)
		io.WriteString(f, currentVersion)
		f.Close()
	}
}

func fetch() string {
	var url string
	if isDevMode {
		url = devURL
	} else {
		url = dxyURL
	}

	req, _ := http.NewRequest("GET", url, strings.NewReader(""))
	myHeaders := map[string]string{
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Host":       "3g.dxy.cn",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0"}
	for k, v := range myHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	res, _ := client.Do(req)

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(res.Body)
	html := string(buf.Bytes())
	return html
}

func prase(html string) dxyDatas {
	peoplesT := regexp.MustCompile("style=\"color: #4169e2\">([0-9]+)</span>").FindAllStringSubmatch(html, 4)
	infoT := regexp.MustCompile("<i class=\"red___3VJ3X\"></i>([\\s\\S]*?)</p>").FindAllStringSubmatch(html, 2)
	info2T := regexp.MustCompile("<i class=\"orange___1FP2_\"></i>([\\S\\s]*?)</p>").FindAllStringSubmatch(html, 10)
	provinceNumberT := regexp.MustCompile("provinceName\":\"" + provinceName + "省\",\"provinceShortName\":\"" + provinceName + "\",\"cityName\":\"\",\"confirmedCount\":([0-9]+),\"suspectedCount\":([0-9]+),\"curedCount\":([0-9]+),\"deadCount\":([0-9]+),").FindStringSubmatch(html)
	cityNumberT := regexp.MustCompile("\"cityName\":\"" + cityName + "\",\"confirmedCount\":([0-9]+),\"suspectedCount\":([0-9]+),\"curedCount\":([0-9]+),\"deadCount\":([0-9]+)").FindStringSubmatch(html)
	d := dxyDatas{
		ddlTime:           strings.Replace(regexp.MustCompile("<p class=\"mapTitle___2QtRg\"><span>([\\S\\s]*?)</span></p>").FindStringSubmatch(html)[1], "\n", "", -1),
		confirmedNumber:   peoplesT[0][1],
		suspectedNumber:   peoplesT[1][1],
		deadNumber:        peoplesT[2][1],
		cureNumber:        peoplesT[3][1],
		susceptiblePeople: strings.Replace(strings.Replace(infoT[0][1], "\n", "", -1), " ", "", -1)[10:],
		incubation:        strings.Replace(infoT[1][1], "\n", "", -1)[8:],
		spreadWay:         strings.Replace(info2T[0][1], "\n", "", -1)[14:],
		isChanged:         strings.Replace(info2T[1][1], "\n", "", -1)[13:],
		diffusion:         strings.Replace(info2T[2][1], "\n", "", -1)[10:],
		provinceNumber:    fmt.Sprintf("%s / %s / %s / %s", provinceNumberT[1], provinceNumberT[2], provinceNumberT[3], provinceNumberT[4]),
		cityNumber:        fmt.Sprintf("%s / %s / %s / %s", cityNumberT[1], cityNumberT[2], cityNumberT[3], cityNumberT[4]),
	}
	return d
}

func main() {
	if isDevMode {
		fmt.Println(prase(fetch()).toString())
	} else {
		cqp.Main()
	}
}

func init() {
	cqp.AppID = appid
	cqp.PrivateMsg = onPrivateMsg
	cqp.GroupMsg = onGroupMsg
	cqp.Enable = onEnable
}

// 定时查询, 当数据更新时发送消息
func onEnable() int32 {
	sendMsg("I am online!", onlineMsgSendStrategy)
	writeLog(fmt.Sprintf("%s", cqp.AppID))
	checkVer()
	d := prase(fetch())
	go func(d *dxyDatas) {
		for {
			current := prase(fetch())
			if d.shouldUpgrade(&current) {
				msgR := d.toStringAfterUpgrade(current)
				writeLog("Upgrade")
				sendMsg(msgR, upgradeSendStrategy)
				d.upgrade(&current)
			}
			time.Sleep(refershInterval * time.Minute)
		}
	}(&d)
	return 0
}

// 私聊发送任何消息都会回复当前情况
func onPrivateMsg(subType, msgID int32, fromQQ int64, msg string, font int32) int32 {
	writeLog(fmt.Sprintf("%d %s", fromQQ, msg))
	msgR := prase(fetch()).toString()
	cqp.SendPrivateMsg(fromQQ, msgR)
	return 0
}

// 群聊中@bot即回复当前情况
func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if strings.Contains(msg, "[CQ:at,qq="+selfQQID+"]") {
		writeLog(fmt.Sprintf("%d %d %s", fromGroup, fromQQ, msg))
		msgR := prase(fetch()).toString()
		cqp.SendGroupMsg(fromGroup, msgR)
	}
	return 0
}
