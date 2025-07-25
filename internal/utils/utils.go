package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// GetEnvOrDefault 获取环境变量，如果不存在则返回默认值
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetPriorityMark 获取优先级标记
func GetPriorityMark(priority *int) string {
	if priority == nil {
		return "⏬"
	}
	switch *priority {
	case 1:
		return "🔽"
	case 3:
		return "🔼"
	case 5:
		return "⏫"
	default:
		return "⏬"
	}
}

// FormatTime 格式化时间字符串
func FormatTime(timeStr, format string) string {
	if timeStr == "" {
		return ""
	}

	// 尝试解析ISO时间格式
	if t := ParseDateTime(timeStr); t != nil {
		return t.Format(format)
	}

	return ""
}

// ParseDateTime 解析时间字符串为time.Time
func ParseDateTime(timeStr string) *time.Time {
	if timeStr == "" {
		return nil
	}

	// 定义东八区时区（北京时间）
	east8Zone := time.FixedZone("CST", 8*3600) // 东八区，UTC+8

	// 支持的时间格式（含时区处理）
	formats := []string{
		time.RFC3339,                   // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05.000-0700", // 毫秒+时区（无冒号）
		"2006-01-02T15:04:05-07:00",    // 带冒号时区
		"2006-01-02T15:04:05.000Z",     // UTC毫秒
		"2006-01-02T15:04:05Z",         // UTC
		"2006-01-02 15:04:05",          // 无时区（默认东八区）
		"2006-01-02",                   // 日期
	}

	for _, format := range formats {
		// 优先尝试带时区解析
		if t, err := time.Parse(format, timeStr); err == nil {
			t = t.In(east8Zone)
			return &t
		}
	}

	return nil
}

// ExtractFrontMatterField 从Front Matter中提取字段值
func ExtractFrontMatterField(content, field string) string {
	pattern := fmt.Sprintf(`%s:\s*(.*?)(?:\n|$)`, regexp.QuoteMeta(field))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// ConvertToBeijingTime 将ISO时间字符串转换为北京时间
func ConvertToBeijingTime(isoTime string) string {
	if isoTime == "" {
		return ""
	}

	t := ParseDateTime(isoTime)
	if t == nil {
		return ""
	}

	// 转换为北京时间 (UTC+8)
	beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
	beijingTime := t.In(beijingLocation)
	return beijingTime.Format("2006-01-02 15:04:05")
}

// GetTodayStamp 获取今天的时间戳
func GetTodayStamp() int {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return int(today.Unix())
}
