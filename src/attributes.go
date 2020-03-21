package main

/*********************************属性字典 start****************************************/
var (
	countsField = [...]string{
		"currentConfirmedCount",
		"confirmedCount",
		"suspectedCount",
		"curedCount",
		"deadCount",
	}

	arrHead = map[string]string{
		"createTimeStr":                "创建时间: ",     // 1579537899000
		"modifyTimeStr":                "更新时间: ",     // 1580141884000
		"infectSource":                 "传染源: ",      // "野生动物，可能为中华菊头蝠"
		"passWay":                      "传播途径: ",     // "未完全掌握，存在人传人、医务人员感染、一定范围社区传播"
		"dailyPic":                     "疫情趋势图: ",    // ["https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg"]
		"dailyPics":                    "疫情趋势图: ",    //
		"summary":                      "汇总: ",       // ""
		"deleted":                      "",           // false
		"countRemark":                  "",           // ""
		"currentConfirmedCount":        "现存确诊: ",     //
		"confirmedCount":               "累计确诊: ",     // 2858
		"suspectedCount":               "境外输入: ",     // 5794
		"seriousCount":                 "现存重症: ",     //
		"curedCount":                   "治愈: ",       // 56
		"deadCount":                    "死亡: ",       // 82
		"foreigncurrentConfirmedCount": "国外现存确诊: ",   //
		"foreignconfirmedCount":        "国外累计确诊: ",   // 2858
		"foreignsuspectedCount":        "国外疑似: ",     // 5794
		"foreigncuredCount":            "国外治愈: ",     // 56
		"foreigndeadCount":             "国外死亡: ",     // 82
		"virus":                        "病毒: ",       // "新型冠状病毒 2019-nCoV"
		"remark1":                      "",           // "易感人群: 暂时不明，病毒存在变异可能"
		"remark2":                      "",           // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
		"remark3":                      "",           // ""
		"remark4":                      "",           // ""
		"remark5":                      "",           // ""
		"note2":                        "",           //
		"note1":                        "",           //
		"note3":                        "",           //
		"generalRemark":                "备注: ",       // "疑似病例数来自国家卫健委数据，目前为全国数据，未分省市自治区等"
		"abroadRemark":                 "",           // ""
		"queryChinaProvinces":          "\n",         //
		"queryGlobalProvinces":         "",           //
		"queryCites":                   "",           //
		"version":                      "\nversion: ", //
		"dxyUrl":                       "\n丁香园: ",    //
		"tencentUrl":                   "腾讯: ",       //
	}

	neededAttributes = [...]string{
		"modifyTimeStr", // 1580141884000
		"currentConfirmedCount",
		"confirmedCount",               // 2858
		"suspectedCount",               // 5794
		"seriousCount",                 //
		"curedCount",                   // 56
		"deadCount",                    // 82
		"foreigncurrentConfirmedCount", //
		"foreignconfirmedCount",        // 2858
		// "foreignsuspectedCount",      // 5794
		"foreigncuredCount",    // 56
		"foreigndeadCount",     // 82
		"queryChinaProvinces",  //
		"queryGlobalProvinces", //
		"queryCites",           //
		// "note2",                //
		// "note1",                //
		// "note3",                //
		// "remark1",              // "易感人群: 暂时不明，病毒存在变异可能"
		// "remark2",              // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
		// "remark3",              // ""
		// "remark4",              // ""
		// "remark5",              // ""
		// //"dailyPics",       // ["https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg"]
		// "dxyUrl",     // 丁香园地址
		// "tencentUrl", // 腾讯新闻地址
		"version", // 版本
	}

	forCheckAttributes = [...]string{
		// "currentConfirmedCount",
		"confirmedCount", // 2858
		"suspectedCount", // 5794
		"seriousCount",   //
		"deadCount",      // 82
		// "curedCount",           // 56
		"foreigncurrentConfirmedCount", //
		"foreignconfirmedCount",        // 2858
		// "foreignsuspectedCount",      // 5794
		"foreigncuredCount",    // 56
		"foreigndeadCount",     // 82
		"queryChinaProvinces",  //
		"queryGlobalProvinces", //
		"queryCites",           //
		// "remark1",              // "易感人群: 暂时不明，病毒存在变异可能"
		// "remark2",              // "潜伏期: 1~14 天均有，平均 10 天，潜伏期内存在传染性"
		// "remark3",              // ""
		// "remark4",              // ""
		// "remark5",              // ""
		// "note1",                // ""
		// "note2",                // ""
		// "note3",                // ""
		// "dailyPics",            // ["https://img1.dxycdn.com/2020/0127/350/3393218957833514634-73.jpg"] (趋势图)
	}
)
