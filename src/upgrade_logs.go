package main

var allVersionUpgradeLog = map[string]string{
	"v2.16.20.28 beta": `A big fuck update.
1. 彻底抛弃代码实现的历史包袱, 统一了所有信息解析方式.appid
2. 更快捷统一的订阅地区管理, 同时支持了国外地区订阅. 同时地区不再需要输入全称.
3. 更强大和完善的命令调用方式, 并抽取和统一了命令解析逻辑.appid
4. 新增命令支持. eg.
   "q 日本 湖北 杨浦": 将返回这三个地区的疫情列表
   "qa 湖北 上海": 将返回湖北及上海及其所有下级城市 / 区的疫情数据列表
   "debug": 调试
   "url": 所有相关监测网址列表
   "空": 返回当前默认数据
5. 增加了代码混乱度, 增加了 bot 崩溃的概率`,

	"v2.13.12.25": `1:.....`,
}