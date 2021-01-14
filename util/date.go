package util

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type JSONTime struct {
	time.Time
}

type JSONDetailTime struct {
	time.Time
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return []byte("\"\""), nil
	}
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02"))
	return []byte(formatted), nil
}

func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t JSONDetailTime) MarshalJSON() ([]byte, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return []byte("\"\""), nil
	}
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

func (t JSONDetailTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *JSONDetailTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONDetailTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

//  ===============

func GetCurrentTimestamp() int64 {
	local, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return time.Now().In(local).Unix()
}

func GetCurrentTimeNanoTimestamp() int64 {
	local, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return time.Now().In(local).UnixNano()
}

func GetCurrentTimeStr() string {
	local, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return time.Now().In(local).Format("2006-01-02 15:04:05")
}

func TimestampToStr(timeNum int64, tpl string) string {
	local, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return time.Unix(timeNum, 0).In(local).Format(tpl)
}

func TimeToStr(t time.Time) string {
	local, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return t.In(local).Format("2006-01-02 15:04:05")
}

func TimeToTimestamp(t time.Time) int64 {
	return StrToTimestamp(TimeToStr(t))
}

func StrToTimestamp(str string) int64 {
	formatTime, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	return formatTime.Unix()
}

func StrToTimeTime(str string) time.Time {
	if len(str) > 10 {
		formatTime, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
		return formatTime
	}
	formatTime, _ := time.ParseInLocation("2006-01-02", str, time.Local)
	return formatTime
}

func GetTodayDate() string {
	year, month, day := time.Now().Date()
	return IntToStr(year) + "_" + IntToStr(int(month)) + "_" + IntToStr(day)
}

// 获取今天剩余的时间戳
func TodayRemainTimestamp() int {
	str := GetCurrentTimeStr()
	expiredTimestamp := StrToTimestamp(str[:10] + " 23:59:59")
	currentTimestamp := GetCurrentTimestamp()
	return int(expiredTimestamp - currentTimestamp)
}

// 获取当天的日期，根据传入的间隔符组合
func GetTodayDateNew(spaceMark string) string {
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	return year + spaceMark + month + spaceMark + day
}

func GetLastDayDate() string {
	year, month, day := time.Now().AddDate(0, 0, -1).Date()
	return IntToStr(year) + "_" + IntToStr(int(month)) + "_" + IntToStr(day)
}

// 获取过去的日期，根据传入的间隔符组合
func GetPastDate(dayCount int, spaceMark string) string {
	pastTime := time.Now().AddDate(0, 0, -dayCount)
	year := pastTime.Format("2006")
	month := pastTime.Format("01")
	day := pastTime.Format("02")
	return year + spaceMark + month + spaceMark + day
}

// 获取日期开始、结束的时间戳
func GetStartAndEndTimestamp(date string) (int64, int64) {
	startTime := StrToTimestamp(date + " 00:00:00")
	endTime := StrToTimestamp(date + " 23:59:59")
	return startTime, endTime
}

//判断时间是当年的第几周
func WeekByDate(t time.Time) int {
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		subDays := yearDay - firstWeekDays
		if subDays%7 == 0 {
			week = subDays/7 + 1
		} else {
			week = subDays/7 + 2
		}
	}
	return week
}

func DateDiffToday(a time.Time) int {
	d := time.Now().Sub(a)
	return int(d.Hours() / 24)
}
