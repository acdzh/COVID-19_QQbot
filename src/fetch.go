package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

/*********************************获取数据 dxyDatas start****************************************/

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
	writeLog("[fetch]")
	return html
}

func prase(html string) dxyDatas {
	sprintf := fmt.Sprintf
	praseSucccess := true
	errorMsg := "网页已改版, 解析失败, 暂停更新. 管理员快来修bug."
	d := make(dxyDatas)

	countryJSONResults := regexp.MustCompile(`getStatisticsService\s=\s({.*?})}`).FindStringSubmatch(html)
	if len(countryJSONResults) == 0 {
		praseSucccess = false
		errorMsg += "\nlen(countryJSONResults) == 0 !"
	} else {
		countryJSON := countryJSONResults[1]
		err := json.Unmarshal([]byte(countryJSON), &d)
		if err != nil {
			praseSucccess = false
			errorMsg += "\nprase json failed !"
		}
	}

	d.dataFmt()

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

	provinceInformationResults2 := regexp.MustCompile(
		sprintf(`"provinceShortName":"%s","confirmedCount":([0-9]+),"suspectedCount":([0-9]+),"curedCount":([0-9]+),"deadCount":([0-9]+),`, provinceShortName2)).FindStringSubmatch(html)

	if len(provinceInformationResults2) == 0 {
		errorMsg += "\nlen(provinceInformationResults2) == 0 !"
		praseSucccess = false
		d["provinceNumber2"] = `%s / %s / %s / %s`
	} else {
		d["provinceNumber2"] = sprintf("%s / %s / %s / %s",
			provinceInformationResults2[1],
			provinceInformationResults2[2],
			provinceInformationResults2[4],
			provinceInformationResults2[3])
	}

	cityInformationResults := regexp.MustCompile(
		sprintf(`{"cityName":"%s","confirmedCount":([0-9]+),"suspectedCount":([0-9]+),"curedCount":([0-9]+),"deadCount":([0-9]+),`, cityName)).FindStringSubmatch(html)

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

	cityInformationResults2 := regexp.MustCompile(
		sprintf(`{"cityName":"%s","confirmedCount":([0-9]+),"suspectedCount":([0-9]+),"curedCount":([0-9]+),"deadCount":([0-9]+),`, cityName2)).FindStringSubmatch(html)

	if len(cityInformationResults2) == 0 {
		errorMsg += "\nlen(cityInformationResults2) == 0"
		praseSucccess = false
		d["cityNumber2"] = `%s / %s / %s / %s`
	} else {
		d["cityNumber2"] = sprintf("%s / %s / %s / %s",
			cityInformationResults2[1],
			cityInformationResults2[2],
			cityInformationResults2[4],
			cityInformationResults2[3])
	}

	if praseSucccess == false {
		if willPraseSuccess {
			if isDevMode {
				fmt.Println(errorMsg)
			} else {
				sendMsg(errorMsg, failedDataSendStrategy)
				writeLog("[prase] error " + errorMsg)

			}
		}
		willPraseSuccess = false
	}

	return d
}
