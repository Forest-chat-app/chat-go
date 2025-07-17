package utils

import "time"

// GetUTCTimestamp 返回UTC标准时间戳
func GetUTCTimestamp() int64 {
	return time.Now().UTC().Unix()
}

// 获取毫秒级UTC时间戳
func GetUTCMillisTimestamp() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

// 把当前时间转换为UTC时间
func GetUTCTime(currentTimestamp int64) int64 {
	return time.Unix(currentTimestamp, 0).UTC().Unix()
}

// 把当前时间转换为毫秒级UTC时间戳
func UnixToUTCMillisTimestamp(currentTimestamp int64) int64 {
	return time.Unix(currentTimestamp, 0).UTC().UnixNano() / int64(time.Millisecond)
}

// FormatTimestamp 将时间戳格式化为指定格式的UTC时间字符串
func FormatTimestamp(timestamp int64, layout string) string {
	if layout == "" {
		layout = time.RFC3339 // 默认格式
	}
	return time.Unix(timestamp, 0).UTC().Format(layout)
}

// ParseTimeToUTC 解析时间字符串为UTC时间戳
func ParseTimeToUTC(timeStr, layout string) (int64, error) {
	if layout == "" {
		layout = time.RFC3339
	}
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.UTC().Unix(), nil
}
