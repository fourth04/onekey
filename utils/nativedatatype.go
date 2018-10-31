package utils

import (
	"encoding/json"
	"reflect"
)

// StructToMap will change struct to map
func StructToMap(inter interface{}) map[string]interface{} {
	param := make(map[string]interface{})

	t := reflect.TypeOf(inter)
	v := reflect.ValueOf(inter)
	for i := 0; i < t.NumField(); i++ {
		param[t.Field(i).Name] = v.Field(i).Interface()
	}
	return param
}

// StructToMapByJson will change struct to map，it can use the struct tag
func StructToMapByJson(inter interface{}) (map[string]interface{}, error) {
	js, _ := json.Marshal(inter)
	var tmpMap map[string]interface{}
	err := json.Unmarshal(js, &tmpMap)
	if err != nil {
		return tmpMap, err
	}
	return tmpMap, nil
}

// Contains slice contain sub
func Contains(slice []string, sub string) bool {
	for _, str := range slice {
		if str == sub {
			return true
		}
	}
	return false
}

// SliceSubtraction returns the elements in a that aren't in b
func SliceSubtraction(sliceA, sliceB []string) []string {
	tmp := map[string]bool{}
	for _, value := range sliceB {
		tmp[value] = true
	}
	rv := []string{}
	for _, value := range sliceA {
		if _, ok := tmp[value]; !ok {
			rv = append(rv, value)
		}
	}
	return rv
}

// 通过两重循环过滤重复元素
func RemoveRepByLoop(slc []string) []string {
	result := []string{} // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

// 通过map主键唯一的特性过滤重复元素
func RemoveRepByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// 元素去重
func RemoveRep(slc []string) []string {
	if len(slc) < 1024 {
		// 切片长度小于1024的时候，循环来过滤
		return RemoveRepByLoop(slc)
	} else {
		// 大于的时候，通过map来过滤
		return RemoveRepByMap(slc)
	}
}

// 将Slice按指定数目分批
func MakeBatch(data []interface{}, batchNum int) [][]interface{} {
	var rv [][]interface{}
	for i := 0; i < len(data); i++ {
		quotient := i / batchNum
		remainder := i % batchNum
		if remainder == batchNum-1 || i == len(data)-1 {
			rv = append(rv, data[quotient*batchNum:quotient*batchNum+remainder+1])
		}
	}
	return rv
}

// 将String Slice按指定数目分批
func MakeStringBatch(data []string, batchNum int) [][]string {
	var rv [][]string
	for i := 0; i < len(data); i++ {
		quotient := i / batchNum
		remainder := i % batchNum
		if remainder == batchNum-1 || i == len(data)-1 {
			rv = append(rv, data[quotient*batchNum:quotient*batchNum+remainder+1])
		}
	}
	return rv
}
