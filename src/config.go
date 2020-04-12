package main

/*****************************自定义数据请在此处修改 strat**********************************/

//运行模式
const (
	runModeDevInConsole      int = 1 // 控制台调试
	runModeDevOnLocalMachine int = 2 // 实机测试
	runModeProduct           int = 3 // 生产环境

	//globalRunMode = runModeDevOnLocalMachine
	//globalRunMode = runModeProduct
	globalRunMode = runModeDevInConsole
)

// 关于更新
const (
	//
	refershInterval     = 5 * 60 // 秒, 主动刷新间隔
	newsRefershInterval = 8 * 60 // 秒, 新闻刷新间隔

	shouldSendAllAfterUpgrade         = false // 当开启时, 下面这句才有效
	shouldSendAllAfterUpgradeInterval = 60    // 分钟, 当间隔时长大于此值时, 发送全部内容, 否则仅发送更新部分

)

var (
	// 自定义查询子区域, 不需要全称, 但必须可以与其他区分
	// 日本, 美国 也是 province, 同样, 亚洲, 非洲 也是 province
	queryChinaProvinceNames  = [...]string{"上海"}
	queryGlobalProvinceNames = [...]string{"美国", "意大利", "西班牙", "德国", "法国", "英国", "伊朗"}
	queryCityNames           = [...]string{}
)

// 基本信息
const (
	appid string = "com.acdzh.dxy" // 务必正确填写

	// bot版本信息
	currentVersion string = "v3.15.21.12" // 当前版本, 每次修改后会进行版本更新推送
	// 版本更新日志, 仅会推送一次
	versionFileName string = "conf/dxy.cfg" // 存储版本号
	logFilePath     string = "data/log/"    // log文件目录 (log会以日期命名
	shouldPushLog   bool   = true           // 是否在更新之后更新版本推送
)

var versionUpgradeLog string = allVersionUpgradeLog[currentVersion]

// how to send msg
const (
	// 消息发送策略模板, 不要修改
	sendToNobody     int = 0 // 不发送给任何类型用户或群组
	sendToUserAndDev int = 1 // 同时发送给普通和管理员用户或群组
	sendTOUserOnly   int = 2 // 仅发送给普通用户或群组
	sendToDevOnly    int = 3 // 仅发送给管理员用户或群组

	// 具体的消息发送策略 (格式为: 10 * 群消息策略 + 私聊消息策略
	onlySendToPrivateDevStrategy int = 10*sendToNobody + sendToDevOnly        // 仅发送给管理员QQ
	onlineMsgSendStrategy        int = 10*sendToNobody + sendToDevOnly        // 上线提醒: 仅私聊发给管理员账号
	firstDataSendStrategy        int = 10*sendToDevOnly + sendToNobody        // 上线后拉取的初始数据: 仅发送到调试qq群
	failedDataSendStrategy       int = 10*sendToNobody + sendToDevOnly        // 出现错误: 仅私聊发送管理员, 并发送给所有群
	versionSendStrategy          int = 10*sendToUserAndDev + sendToDevOnly    // 版本日志: 发送给所有群, 但私聊仅发送给管理员
	upgradeSendStrategy          int = 10*sendToNobody + sendToUserAndDev     // 数据更新: 仅发给管理员用户
	newsUpgradeSendStrategy      int = 10*sendToUserAndDev + sendToUserAndDev // 新闻更新: 发送给所有群和用户
)

// qqGroup & qqID
var (
	selfQQID       string = "1472745738"          // bot自己的qq号
	userQQGroupIDs        = [...]int64{}          // 普通用户qq群数组
	devQQGroupIDs         = [...]int64{584405782} // 开发者调试用qq群数组
	userQQIds             = [...]int64{}          // 普通用户订阅qq号数组
	devQQIds              = []int64{1069436872}   // 开发者qq号数组
)

// url
const (
	dxyURL     string = "https://ncov.dxy.cn/ncovh5/view/pneumonia"             // 数据来源url
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

// 全局变量
var willPraseSuccess bool = true            // 标识是否会解析失败
var lastSendAllAfterUpgradeTime float64 = 0 // 上一次更新推送全部项是什么时候(unix时间戳 / ms)
var lastNewsTimeStamp int64 = 0             //1584339618    //1583336893    // 上一条news的日期 (发布更新之前不要忘了手动改一下
