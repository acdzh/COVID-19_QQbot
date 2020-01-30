package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/CoolQ-Golang-SDK/v2/cqp"
)

/*****************************自定义数据请在此处修改**********************************/
var willPraseSuccess bool = true

const (
	// 基本信息
	appid string = "com.acdzh.dxy"

	//调试模式
	isDevMode bool = false

	// 主动刷新间隔
	refershInterval = 5 // 分钟

	// 自定义查询子区域 (未对所有地市进行匹配, 如果失败请自行修改正则
	provinceName      string = "山东省"
	provinceShortName string = "山东"
	cityName          string = "菏泽"

	// bot版本信息
	currentVersion string = "v1.30.10.59" // 当前版本, 每次修改后会进行版本更新推送
	// 版本更新日志, 仅会推送一次
	versionUpgradeLog string = `1. 关闭地图更新提醒`
	versionFileName   string = "conf/dxy.cfg" // 存储版本号
	logFilePath       string = "data/log/"    // log文件目录 (log会以日期命名
	shouldPushLog     bool   = true           // 是否在每次更新之后更新版本推送

	// url
	dxyURL     string = "https://3g.dxy.cn/newh5/view/pneumonia"                // 数据来源url
	baiduURL   string = "https://voice.baidu.com/act/newpneumonia/newpneumonia" // 地图来源uurl
	tencentURL string = "https://news.qq.com/zt2020/page/feiyan.htm"
	devURL     string = "http://127.0.0.1:5500/index.html" // 本地调试url
	urlList    string = `其他监测网址:
凤凰网: https://news.ifeng.com/c/special/7tPlDSzDgVk
新浪: https://news.sina.cn/zt_d/yiqing0121
百度: https://voice.baidu.com/act/newpneumonia/newpneumonia
搜狗: https://123.sogou.com/zhuanti/pneumonia.html
知乎: https://www.zhihu.com/special/19681091
网易: https://news.163.com/special/epidemic/
头条: https://i.snssdk.com/feoffline/hot_list/template/hot_list/forum.html?forum_id=1656388947394568
夸克: https://broccoli.uc.cn/apps/pneumonia/routes/index`
)

var (
	// qqGroup & qqID
	selfQQID       string = "1472745738"                     // bot自己的qq号
	userQQGroupIDs        = [...]int64{854378285, 361684286} // 普通用户qq群数组
	devQQGroupIDs         = [...]int64{584405782}            // 开发者调试用qq群数组
	userQQIds             = [...]int64{}                     // 普通用户订阅qq号数组
	devQQIds              = []int64{1069436872}              // 开发者qq号数组
)

const (
	// 消息发送策略模板, 不要修改
	sendToNobody     int = 0 // 不发送给任何类型用户或群组
	sendToUserAndDev int = 1 // 同时发送给普通和管理员用户或群组
	sendTOUserOnly   int = 2 // 仅发送给普通用户或群组
	sendToDevOnly    int = 3 // 仅发送给管理员用户或群组

	// 具体的消息发送策略 (格式为: 10 * 群消息策略 + 私聊消息策略
	onlySendToPrivateDevStrategy int = 10*sendToNobody + sendToDevOnly
	onlineMsgSendStrategy        int = 10*sendToNobody + sendToDevOnly        // 上线提醒: 仅私聊发给管理员账号
	firstDataSendStrategy        int = 10*sendToDevOnly + sendToNobody        // 上线后拉取的初始数据: 仅发送到调试qq群
	failedDataSendStrategy       int = 10*sendToUserAndDev + sendToDevOnly    // 出现错误: 仅私聊发送管理员, 并发送给所有群
	versionSendStrategy          int = 10*sendToUserAndDev + sendToDevOnly    // 版本日志: 发送给所有群, 但私聊仅发送给管理员
	upgradeSendStrategy          int = 10*sendToUserAndDev + sendToUserAndDev // 数据更新: 发送给所有群和用户
)

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

func timeStampToString(t string) string {
	if t == "" {
		return "NaN"
	}
	i, _ := strconv.ParseInt(t, 10, 64)
	return time.Unix(i/1000, 0).Format("2006-01-02 15:04:05 (北京时间)")
}

var arrHead = map[string]string{
	"createTime":     "创建时间: ",            // 1579537899000
	"modifyTime":     "更新时间: ",            // 1580141884000
	"infectSource":   "传染源: ",             // "野生动物，可能为中华菊头蝠"
	"passWay":        "传播途径: ",            // "未完全掌握，存在人传人、医务人员感染、一定范围社区传播"
	"imgUrl":         "\n疫情地图: ",          // "https://img1.dxycdn.com/2020/0123/733/3392575782185696736-73.jpg"
	"dailyPic":       "疫情趋势图: ",           // "https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg"
	"summary":        "汇总: ",              // ""
	"deleted":        "",                  // false
	"countRemark":    "",                  // ""
	"confirmedCount": "确诊: ",              // 2858
	"suspectedCount": "疑似: ",              // 5794
	"curedCount":     "治愈: ",              // 56
	"deadCount":      "死亡: ",              // 82
	"virus":          "病毒: ",              // "新型冠状病毒 2019-nCoV"
	"remark1":        "",                  // "易感人群: 暂时不明，病毒存在变异可能"
	"remark2":        "",                  // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
	"remark3":        "",                  // ""
	"remark4":        "",                  // ""
	"remark5":        "",                  // ""
	"generalRemark":  "备注: ",              // "疑似病例数来自国家卫健委数据，目前为全国数据，未分省市自治区等"
	"abroadRemark":   "",                  // ""
	"provinceNumber": provinceName + ": ", // 1 / 2 / 3 / 4
	"cityNumber":     cityName + "市: ",    // 1 / 2 / 3 / 4
	"version":        "\nbot当前版本: ",
	"dxyUrl":         "\n丁香园: ",
	"tencentUrl":     "腾讯: ",
}

var allAttributes = [...]string{
	"createTime",     // 1579537899000
	"modifyTime",     // 1580141884000
	"infectSource",   // "野生动物，可能为中华菊头蝠"
	"passWay",        // "未完全掌握，存在人传人、医务人员感染、一定范围社区传播"
	"imgUrl",         // "https://img1.dxycdn.com/2020/0123/733/3392575782185696736-73.jpg"
	"dailyPic",       // "https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg"
	"summary",        // ""
	"deleted",        // false
	"countRemark",    // ""
	"confirmedCount", // 2858
	"suspectedCount", // 5794
	"curedCount",     // 56
	"deadCount",      // 82
	"virus",          // "新型冠状病毒 2019-nCoV"
	"remark1",        // "易感人群: 暂时不明，病毒存在变异可能"
	"remark2",        // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
	"remark3",        // ""
	"remark4",        // ""
	"remark5",        //
	"generalRemark",  // "疑似病例数来自国家卫健委数据，目前为全国数据，未分省市自治区等"
	"abroadRemark",   // ""
}

var otherAttributes = [...]string{
	"provinceNumber", // 省份的数据
	"cityNumber",     // 市的数据
}

var neededAttributes = [...]string{
	"modifyTime",     // 1580141884000
	"confirmedCount", // 2858
	"suspectedCount", // 5794
	"deadCount",      // 82
	"curedCount",     // 56
	"provinceNumber", // 省份的数据
	"cityNumber",     // 市的数据
	"infectSource",   // "野生动物，可能为中华菊头蝠"
	"virus",          // "新型冠状病毒 2019-nCoV"
	"passWay",        // "未完全掌握，存在人传人、医务人员感染、一定范围社区传播"
	"remark1",        // "易感人群: 暂时不明，病毒存在变异可能"
	"remark2",        // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
	"remark3",        // ""
	"remark4",        // ""
	"remark5",        // ""
	"imgUrl",         // "https://img1.dxycdn.com/2020/0123/733/3392575782185696736-73.jpg"
	"dailyPic",       // "https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg"
	"dxyUrl",         // 丁香园地址
	"tencentUrl",     // 腾讯新闻地址
	"version",        // 版本
}

var forCheckAttributes = [...]string{
	"confirmedCount", // 2858
	"suspectedCount", // 5794
	"deadCount",      // 82
	"curedCount",     // 56
	"provinceNumber", // 省份的数据
	"cityNumber",     // 市的数据
	"infectSource",   // "野生动物，可能为中华菊头蝠"
	"virus",          // "新型冠状病毒 2019-nCoV"
	"passWay",        // "未完全掌握，存在人传人、医务人员感染、一定范围社区传播"
	"remark1",        // "易感人群: 暂时不明，病毒存在变异可能"
	"remark2",        // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
	"remark3",        // ""
	"remark4",        // ""
	"remark5",        // ""
	"dailyPic",       // "https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg" (趋势图)
}

type dxyDatas map[string]string

func (d dxyDatas) toString() string {
	s := ""
	for _, arr := range neededAttributes {
		lineHead := arrHead[arr]
		lineBody := d[arr]
		if lineHead != "" || lineBody != "" {
			s += (lineHead + lineBody + "\n")
		}
	}
	return strings.TrimRight(s, "\n")
}

func (d dxyDatas) toStringBeforeUpgrade(new dxyDatas) string {
	upgradeFormat := func(a, b string) string {
		if a == b {
			return a
		}
		return fmt.Sprintf("%s -> %s (已更新)", a, b)
	}

	s := ""
	for _, arr := range neededAttributes {
		lineHead := arrHead[arr]
		lineBody := upgradeFormat(d[arr], new[arr])
		if lineHead != "" || lineBody != "" {
			s += (lineHead + lineBody + "\n")
		}
	}
	return strings.TrimRight(s, "\n")
}

func (d dxyDatas) shouldUpgrade(new dxyDatas) bool {
	for _, arr := range forCheckAttributes {
		if d[arr] != new[arr] {
			return true
		}
	}
	return false
}

func (d dxyDatas) upgrade(new dxyDatas) {
	for _, arr := range allAttributes {
		d[arr] = new[arr]
	}
	for _, arr := range otherAttributes {
		d[arr] = new[arr]
	}
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
		f, _ := os.OpenFile(versionFileName, os.O_WRONLY|os.O_TRUNC, 0666)
		io.WriteString(f, currentVersion)
		f.Close()
	}
}

func fetch() string {
	if !willPraseSuccess {
		return ""
	}
	url := dxyURL

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

func fetchMap() string {
	if !willPraseSuccess {
		return ""
	}
	url := baiduURL

	req, _ := http.NewRequest("GET", url, strings.NewReader(""))
	myHeaders := map[string]string{
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0"}
	for k, v := range myHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	res, _ := client.Do(req)

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(res.Body)
	html := string(buf.Bytes())

	result := regexp.MustCompile(`"mapSrc":"https:\\/\\/mms-res.cdn.bcebos.com\\/mms-res\\/voicefe\\/captain\\/images\\/(.*?).png`).FindString(html)
	if len(result) == 0 {
		return dxyURL
	}
	return strings.Replace(result[10:], "\\", "", -1)
}

func prase(html string) dxyDatas {
	sprintf := fmt.Sprintf
	praseSucccess := true
	errorMsg := "网页已改版, 解析失败. 管理员快来修bug."
	d := make(dxyDatas)

	contryInformationResults := regexp.MustCompile(`{"id":([0-9]+),"createTime":([0-9]+),"modifyTime":([0-9]+),"infectSource":"(.*?)","passWay":"(.*?)","imgUrl":"(.*?)","dailyPic":"(.*?)","summary":"(.*?)","deleted":([\S]+),"countRemark":"(.*?)","confirmedCount":([0-9]+),"suspectedCount":([0-9]+),"curedCount":([0-9]+),"deadCount":([0-9]+),"virus":"(.*?)","remark1":"(.*?)","remark2":"(.*?)","remark3":"(.*?)","remark4":"(.*?)","remark5":"(.*?)","generalRemark":"(.*?)","abroadRemark":"(.*?)"}`).FindStringSubmatch(html)

	if len(contryInformationResults) == 0 {
		praseSucccess = false
		errorMsg += "\nlen(contryInformationResults) == 0 !"
		for _, arr := range allAttributes {
			d[arr] = ""
		}
	} else {
		for index, arr := range allAttributes {
			d[arr] = contryInformationResults[index+2]
		}
	}
	d["createTime"] = timeStampToString(d["createTime"])
	d["modifyTime"] = timeStampToString(d["modifyTime"])
	d["dxyUrl"] = dxyURL
	d["tencentUrl"] = tencentURL
	d["imgUrl"] = fetchMap()
	d["version"] = currentVersion

	provinceInformationResults := regexp.MustCompile(
		sprintf(`"provinceShortName":"%s","confirmedCount":([0-9]+),"suspectedCount":([0-9]+),"curedCount":([0-9]+),"deadCount":([0-9]+),`, provinceShortName)).FindStringSubmatch(html)

	if len(provinceInformationResults) == 0 {
		errorMsg += "\nlen(provinceInformationResults) == 0 !"
		praseSucccess = false
		d["provinceNumber"] = `%s / %s / %s / %s`
	} else {
		d["provinceNumber"] = sprintf("%s / %s / %s / %s",
			provinceInformationResults[1],
			provinceInformationResults[2],
			provinceInformationResults[4],
			provinceInformationResults[3])
	}

	cityInformationResults := regexp.MustCompile(
		sprintf(`{"cityName":"%s","confirmedCount":([0-9]+),"suspectedCount":([0-9]+),"curedCount":([0-9]+),"deadCount":([0-9]+)}`, cityName)).FindStringSubmatch(html)

	if len(cityInformationResults) == 0 {
		errorMsg += "\nlen(cityInformationResults) == 0"
		praseSucccess = false
		d["cityNumber"] = `%s / %s / %s / %s`
	} else {
		d["cityNumber"] = sprintf("%s / %s / %s / %s",
			cityInformationResults[1],
			cityInformationResults[2],
			cityInformationResults[4],
			cityInformationResults[3])
	}

	if praseSucccess == false {
		if willPraseSuccess {
			if isDevMode {
				fmt.Println(errorMsg)
			} else {
				sendMsg(errorMsg, failedDataSendStrategy)

			}
		}
		willPraseSuccess = false
	}

	return d
}

func main() {
	if isDevMode {
		html := fetch()
		d := prase(html)
		dd := prase(html)
		dd["deadCount"] = "9843"
		fmt.Println(d)
		fmt.Println()
		fmt.Println(d.toString())
		fmt.Println()
		fmt.Println(d.toStringBeforeUpgrade(dd))
		d.upgrade(dd)
		fmt.Println()
		fmt.Println(d.toString())
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
	go func(d dxyDatas) {
		for {
			if willPraseSuccess {
				current := prase(fetch())
				if d.shouldUpgrade(current) {
					msgR := d.toStringBeforeUpgrade(current)
					writeLog("Upgrade")
					sendMsg(msgR, upgradeSendStrategy)
					d.upgrade(current)
				}
			}
			time.Sleep(refershInterval * time.Minute)
		}
	}(d)
	return 0
}

// 私聊发送任何消息都会回复当前情况
func onPrivateMsg(subType, msgID int32, fromQQ int64, msg string, font int32) int32 {
	writeLog(fmt.Sprintf("%d %s", fromQQ, msg))
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
		writeLog(fmt.Sprintf("%d %d %s", fromGroup, fromQQ, msg))
		if strings.Contains(msg, "url") {
			cqp.SendGroupMsg(fromGroup, urlList)
			return 0
		}
		msgR := prase(fetch()).toString()
		cqp.SendGroupMsg(fromGroup, msgR)
	}
	return 0
}
