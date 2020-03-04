package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func fetchNews() string {
	writeLog("[fetchNews] start")
	req, _ := http.NewRequest("GET", "https://ncov-rss.qgis.me/api/messages?limit=1", strings.NewReader(""))
	myHeaders := map[string]string{
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0"}
	for k, v := range myHeaders {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	writeLog("[fetchNews] get response done")

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(res.Body)
	html := string(buf.Bytes())
	writeLog("[fetchNews] done")
	return html
}

func timeParseForNews(timeStr string) (timestamp int64, timeStrChina string) {
	withNanos := "2006-01-02T15:04:05-07:00"
	t1, _ := time.Parse(
		withNanos,
		timeStr)
	t1 = t1.In(time.Local)
	return t1.Unix(), t1.Format("2006-01-02 15:04:05 (北京时间)")

}

func parseNews(html string) (isUpdated bool, message string, err error) {
	writeLog("[parseNews] begin")
	var j interface{}
	err = json.Unmarshal([]byte(html), &j)
	if err != nil {
		return false, "parse failed.", fmt.Errorf("parse failed.")
	}
	lastNews := j.(map[string]interface{})["messages"].([]interface{})[0].(map[string]interface{})
	timeStampCST, timeStr := timeParseForNews(lastNews["date"].(string))
	if timeStampCST <= lastNewsTimeStamp {
		writeLog("[parseNews] fail done")
		return false, "", nil
	}
	lastNewsTimeStamp = timeStampCST
	message = lastNews["message"].(string) + "\n"
	for _, entity := range (lastNews["entities"]).([]interface{}) {
		entityJSON := entity.(map[string]interface{})
		if entityJSON["_"].(string) == "MessageEntityTextUrl" {
			offset := (int)(entityJSON["offset"].(float64))
			length := (int)(entityJSON["length"].(float64))
			url := entityJSON["url"].(string)
			message += fmt.Sprintf("\n%s: %s", string([]rune(message)[offset:offset+length]), url)
		}
	}
	message += fmt.Sprintf("\n\n更新时间: %s", timeStr)
	writeLog("[parseNews] success done")
	return true, message, nil
}
