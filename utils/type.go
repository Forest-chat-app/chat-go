// utils/type_convert.go
package utils

import (
	"fmt"
	"strconv"
)

// GetStringValue 从map中获取字符串值，如果不存在或类型不匹配则返回空字符串
func GetStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		// 尝试将其他类型转换为字符串
		return ToString(val)
	}
	return ""
}

// GetIntValue 从map中获取整数值，如果不存在或无法转换则返回0
func GetIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		return ToInt(val)
	}
	return 0
}

// GetFloatValue 从map中获取浮点值，如果不存在或无法转换则返回0
func GetFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		return ToFloat64(val)
	}
	return 0
}

// GetBoolValue 从map中获取布尔值，如果不存在或无法转换则返回false
func GetBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
		// 尝试字符串转换
		if s, ok := val.(string); ok {
			b, _ := strconv.ParseBool(s)
			return b
		}
		// 数字：非零为true
		if i, ok := val.(int); ok {
			return i != 0
		}
		if f, ok := val.(float64); ok {
			return f != 0
		}
	}
	return false
}

// ToInt 将任意类型转换为int
func ToInt(val interface{}) int {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		// 尝试转换字符串为整数
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		// 尝试转换字符串为浮点数再转整数
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int(f)
		}
	}
	return 0
}

// ToFloat64 将任意类型转换为float64
func ToFloat64(val interface{}) float64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		// 尝试转换字符串为浮点数
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

// ToString 将任意类型转换为字符串
func ToString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	}

	// 尝试使用fmt.Sprint处理其他类型
	return fmt.Sprintf("%v", val)
}

// GetMapValue 从map中获取嵌套的map值
func GetMapValue(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return nil
}
