package main

import "fmt"

func consoleTest() {
	// html := fetch()
	// d := prase(html)
	// lastSendAllAfterUpgradeTime = d["modifyTime"].(float64)
	// writeLog(fmt.Sprintf("%v\n", lastSendAllAfterUpgradeTime))
	// dd := prase(html)
	// dd["deadCount"] = "9843 (较昨日 +59)"
	// dd["modifyTime"] = 1580722478000.0
	// for k, v := range d {
	// 	writeLog(fmt.Sprintf("%v, %v", k, v))
	// }

	// writeLog("")
	// writeLog(d.toString())
	// writeLog("")
	// writeLog(d.toStringBeforeUpgrade(dd))
	// d.upgrade(dd)
	// writeLog("")
	// writeLog(d.toString())
	s := fetchNews()
	_, tt, _ := parseNews(s)
	fmt.Println(tt)
	for _, i := range tt {
		fmt.Println(i)
	}
}
