package util

import (
	"encoding/json"
	"strconv"
)

// ===== int to string ======
func IntToStr(num int) string {
	return strconv.Itoa(num)
}

func UIntToStr(num uint) string {
	return IntToStr(int(num))
}

func Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}

// ===== string to int =======
func StrToInt(str string) int {
	ret, _ := strconv.Atoi(str)
	return ret
}

func StrToUInt(str string) uint {
	return uint(StrToInt(str))
}

func StrToInt8(str string) int8 {
	ret, _ := strconv.ParseInt(str, 10, 8)
	return int8(ret)
}

func StrToInt64(str string) int64 {
	ret, _ := strconv.ParseInt(str, 10, 64)
	return int64(ret)
}

func InterfaceToMap(obj interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	bytes, _ := json.Marshal(obj)
	err := json.Unmarshal(bytes, &ret)
	if err != nil {
		panic(err)
	}
	return ret
}
