package main

import (
	"bytes"
	"net/http"
	"strings"
)

/*********************************获取数据 dxyDatas start****************************************/

func fetch() string {
	if !willPraseSuccess {
		return ""
	}
	var url string
	if globalRunMode == runModeDevOnLocalMachine {
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
	writeLog("[fetch]")
	return html
}
