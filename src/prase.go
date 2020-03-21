package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

/*********************************解析数据 dxyDatas start****************************************/

func extractDataFromHTML(html, field string) (interface{}, error) {
	JSONResult := regexp.MustCompile(field + `\s=\s([\[|{].*?[}|\]])}catch\(e\){}`).FindStringSubmatch(html)
	if len(JSONResult) == 0 {
		return nil, fmt.Errorf("can't find datas of field %s", field)
	}
	var j interface{}
	err := json.Unmarshal([]byte(JSONResult[1]), &j)
	if err != nil {
		return j, fmt.Errorf("parse field %s failed", field)
	}
	return j, nil
}

// 国外情况列表
func htmlGetListByCountryTypeService2(html string) ([]interface{}, error) {
	j, err := extractDataFromHTML(html, "getListByCountryTypeService2true")
	if j == nil {
		fuckedUp("获取国外列表 error")
		return nil, fmt.Errorf("获取国外列表 error")
	}
	return j.([]interface{}), err
}

// 国内全部情况列表
func htmlGetAreaStat(html string) ([]interface{}, error) {
	j, err := extractDataFromHTML(html, "getAreaStat")
	if j == nil {
		fuckedUp("获取国内列表 error")
		return nil, fmt.Errorf("获取国内列表 error")
	}
	return j.([]interface{}), err
}

func htmlGetAllProvinceAndCity(html string) ([]interface{}, []interface{}, error) {
	j1, err := htmlGetListByCountryTypeService2(html)
	if err != nil {
		return nil, nil, err
	}
	j3, err := htmlGetAreaStat(html)
	return j1, j3, err
}

// 国内总体情况字典
func htmlGetStatisticsService(html string) (dxyDatas, error) {
	j, err := extractDataFromHTML(html, "getStatisticsService")
	return dxyDatas(j.(map[string]interface{})), err
}

// 时间线列表
func htmlGetTimelineService(html string) ([]interface{}, error) {
	j, err := extractDataFromHTML(html, "getTimelineService")
	return j.([]interface{}), err
}

func getDatasOfProvinceFromOneList(datas []interface{}, queryProvinceName string) (map[string]string, string, error) {
	for _, v := range datas {
		provinceData := v.(map[string]interface{})
		provinceName := provinceData["provinceName"].(string)
		if strings.Contains(provinceName, queryProvinceName) {
			peoplesCount := make(map[string]string)
			for _, field := range countsField {
				peoplesCount[field] = itos(provinceData[field])
			}
			return peoplesCount, provinceName, nil
		}
	}
	return nil, "%s", fmt.Errorf("not find %s", queryProvinceName)
}

func getDatasOfGlobalProvince(globalDatas, chinaDatas []interface{}, queryProvinceName string) (map[string]string, string, error) {
	counts, name, err := getDatasOfProvinceFromOneList(globalDatas, queryProvinceName)
	if err != nil {
		counts, name, err = getDatasOfProvinceFromOneList(chinaDatas, queryProvinceName)
	}
	return counts, name, err
}

func getDatasOfCity(datas []interface{}, queryCityName string) (map[string]string, string, error) {
	for _, v := range datas {
		for _, vv := range v.(map[string]interface{})["cities"].([]interface{}) {
			cityData := vv.(map[string]interface{})
			cityName := cityData["cityName"].(string)
			if strings.Contains(cityName, queryCityName) {
				peoplesCount := make(map[string]string)
				for _, field := range countsField {
					peoplesCount[field] = itos(cityData[field])
				}
				return peoplesCount, cityName, nil
			}
		}
	}
	return nil, "%s", fmt.Errorf("not find %s", queryCityName)
}

func getDataStrsOfCitesOfAProvince(datas []interface{}, queryProvinceName string) (string, error) {
	for _, v := range datas {
		provinceData := v.(map[string]interface{})
		provinceName := provinceData["provinceName"].(string)
		if !strings.Contains(provinceName, queryProvinceName) {
			continue
		}
		peoplesCount := make(map[string]string)
		for _, field := range countsField {
			peoplesCount[field] = itos(provinceData[field])
		}
		results := fmt.Sprintf("%s: %s", provinceName, peopleCountsToString(peoplesCount))
		for _, vv := range provinceData["cities"].([]interface{}) {
			cityData := vv.(map[string]interface{})
			cityName := cityData["cityName"].(string)
			peoplesCount := make(map[string]string)
			for _, field := range countsField {
				peoplesCount[field] = itos(cityData[field])
			}
			results += fmt.Sprintf("\n%s: %s", cityName, peopleCountsToString(peoplesCount))
		}
		return results, nil
	}
	return "%s", fmt.Errorf("not find %s", queryProvinceName)
}

func getDatasOfProvinceOrCity(globalDatas, areaDatas []interface{}, queryName string) (map[string]string, string, error) {
	counts, name, err := getDatasOfProvinceFromOneList(globalDatas, queryName)
	if err != nil {
		counts, name, err = getDatasOfProvinceFromOneList(areaDatas, queryName)
	}
	if err != nil {
		counts, name, err = getDatasOfCity(areaDatas, queryName)
	}
	return counts, name, err
}

func prase(html string) dxyDatas {
	sprintf := fmt.Sprintf
	praseSucccess := true
	errorMsg := "error."

	d, err := htmlGetStatisticsService(html)
	if err != nil {
		praseSucccess = false
		errorMsg += sprintf("\nprase json failed: %v.", err)
	}
	d.dataFmt()

	allGlobalProvinceDatas, err := htmlGetListByCountryTypeService2(html)
	if err != nil {
		praseSucccess = false
		errorMsg += sprintf("\nquery global province failed: %v.", err)
	} else {
		queryGlobalprovinceDatasStr := ""
		for _, queryProvinceName := range queryGlobalProvinceNames {
			d, name, err2 := getDatasOfProvinceFromOneList(allGlobalProvinceDatas, queryProvinceName)
			if err2 != nil {
				praseSucccess = false
				errorMsg += sprintf("\n%v", err2)
			}
			queryGlobalprovinceDatasStr += sprintf("%s: %s\n", name, peopleCountsToString(d))
		}
		d["queryGlobalProvinces"] = strings.TrimRight(queryGlobalprovinceDatasStr, "\n")
	}

	allCityDatas, err := htmlGetAreaStat(html)
	if err != nil {
		praseSucccess = false
		errorMsg += sprintf("\nquery city failed: %v.", err)
	} else {
		queryCityDatasStr := ""
		for _, queryCityName := range queryCityNames {
			d, name, err2 := getDatasOfCity(allCityDatas, queryCityName)
			if err2 != nil {
				praseSucccess = false
				errorMsg += sprintf("\n%v", err2)
			}
			queryCityDatasStr += sprintf("%s: %s\n", name, peopleCountsToString(d))
		}
		d["queryCites"] = strings.TrimRight(queryCityDatasStr, "\n")

		queryChinaprovinceDatasStr := ""
		for _, queryProvinceName := range queryChinaProvinceNames {
			d, name, err2 := getDatasOfProvinceFromOneList(allCityDatas, queryProvinceName)
			if err2 != nil {
				praseSucccess = false
				errorMsg += sprintf("\n%v", err2)
			}
			queryChinaprovinceDatasStr += sprintf("%s: %s\n", name, peopleCountsToString(d))
		}
		d["queryChinaProvinces"] = strings.TrimRight(queryChinaprovinceDatasStr, "\n")
	}

	if praseSucccess == false {
		fuckedUp(errorMsg)
	}

	return d
}
