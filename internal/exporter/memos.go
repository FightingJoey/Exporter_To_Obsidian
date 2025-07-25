package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"exporter-to-obsidian/internal/types"
	"exporter-to-obsidian/internal/utils"
)

// MemosExporter Memos导出器
type MemosExporter struct {
	records     []types.MemosRecord
	outputDir   string
	calendarDir string
	dailyDir    string
	weeklyDir   string
}

// NewMemosExporter 创建新的Memos导出器
func NewMemosExporter(records []types.MemosRecord, outputDir string) *MemosExporter {
	if outputDir == "" {
		outputDir = utils.GetEnvOrDefault("OUTPUT_DIR", ".")
	}

	calendarDir := filepath.Join(outputDir, utils.GetEnvOrDefault("MEMOS_DIR", "Memos"))

	exporter := &MemosExporter{
		records:     records,
		outputDir:   outputDir,
		calendarDir: calendarDir,
		dailyDir:    filepath.Join(calendarDir, "1.Daily"),
		weeklyDir:   filepath.Join(calendarDir, "2.Weekly"),
	}

	// 确保目录存在
	dirs := []string{
		exporter.calendarDir,
		exporter.dailyDir,
		exporter.weeklyDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", dir, err)
		}
	}

	return exporter
}

// ExportDailySummary 导出每日Memos摘要
func (e *MemosExporter) ExportDailySummary(date time.Time) error {
	// 设置日期范围
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endDate := startDate.Add(24*time.Hour - time.Second)

	// 获取当日的Memos记录
	dailyRecords := e.getRecordsInDateRange(startDate, endDate)

	// 创建文件名
	filename := fmt.Sprintf("%s-Memos.md", date.Format("2006-01-02"))
	filepath := filepath.Join(e.dailyDir, filename)

	// 准备文件内容
	content := e.getSummaryFrontMatter()
	content += fmt.Sprintf("# %s Memos摘要\n\n", date.Format("2006-01-02"))

	if len(dailyRecords) > 0 {
		// 按时间排序（最新的在前）
		sort.Slice(dailyRecords, func(i, j int) bool {
			timeI := int64(0)
			if dailyRecords[i].CreatedTs != nil {
				timeI = *dailyRecords[i].CreatedTs
			}
			timeJ := int64(0)
			if dailyRecords[j].CreatedTs != nil {
				timeJ = *dailyRecords[j].CreatedTs
			}
			return timeI > timeJ
		})

		for _, record := range dailyRecords {
			content += e.formatMemosRecord(record)
		}
	} else {
		content += "今日没有Memos记录。\n"
	}

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入每日Memos摘要失败: %v", err)
	}

	fmt.Printf("已创建每日Memos摘要：%s\n", filename)
	return nil
}

// ExportWeeklySummary 导出每周Memos摘要
func (e *MemosExporter) ExportWeeklySummary(date time.Time) error {
	// 获取周的开始和结束日期（周一到周日）
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}
	startOfWeek := date.AddDate(0, 0, -(weekday - 1))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// 设置时间范围
	startDate := time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, time.Local)
	endDate := time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 999999999, time.Local)

	// 获取该周的Memos记录
	weeklyRecords := e.getRecordsInDateRange(startDate, endDate)

	// 创建文件名
	filename := fmt.Sprintf("%s-Week-Memos.md", startOfWeek.Format("2006-01-02"))
	filepath := filepath.Join(e.weeklyDir, filename)

	// 准备文件内容
	content := e.getSummaryFrontMatter()
	content += fmt.Sprintf("# %s 至 %s 周Memos摘要\n\n", startOfWeek.Format("2006-01-02"), endOfWeek.Format("2006-01-02"))

	if len(weeklyRecords) > 0 {
		// 按天聚合记录
		recordsByDay := make(map[string][]types.MemosRecord)
		for _, record := range weeklyRecords {
			recordDate := e.getRecordDate(record)
			if recordDate != "" {
				recordsByDay[recordDate] = append(recordsByDay[recordDate], record)
			}
		}

		// 为每天生成内容
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dayStr := d.Format("2006-01-02")
			dayRecords := recordsByDay[dayStr]

			content += fmt.Sprintf("## %s\n\n", dayStr)

			if len(dayRecords) > 0 {
				// 按时间排序（最新的在前）
				sort.Slice(dayRecords, func(i, j int) bool {
					timeI := int64(0)
					if dayRecords[i].CreatedTs != nil {
						timeI = *dayRecords[i].CreatedTs
					}
					timeJ := int64(0)
					if dayRecords[j].CreatedTs != nil {
						timeJ = *dayRecords[j].CreatedTs
					}
					return timeI > timeJ
				})

				for _, record := range dayRecords {
					content += e.formatMemosRecord(record)
				}
			} else {
				content += "无Memos记录\n"
			}
			content += "\n"
		}
	} else {
		content += "本周没有Memos记录。\n"
	}

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入每周Memos摘要失败: %v", err)
	}

	fmt.Printf("已创建每周Memos摘要：%s\n", filename)
	return nil
}

// getRecordsInDateRange 获取指定日期范围内的Memos记录
func (e *MemosExporter) getRecordsInDateRange(startDate, endDate time.Time) []types.MemosRecord {
	var records []types.MemosRecord

	for _, record := range e.records {
		if record.CreatedTs != nil {
			recordTime := time.Unix(*record.CreatedTs, 0)
			if !recordTime.Before(startDate) && !recordTime.After(endDate) {
				records = append(records, record)
			}
		}
	}

	return records
}

// getRecordDate 获取记录的日期字符串
func (e *MemosExporter) getRecordDate(record types.MemosRecord) string {
	if record.CreatedTs != nil {
		recordTime := time.Unix(*record.CreatedTs, 0)
		return recordTime.Format("2006-01-02")
	}
	return ""
}

// formatMemosRecord 格式化Memos记录
func (e *MemosExporter) formatMemosRecord(record types.MemosRecord) string {
	var content strings.Builder

	// 添加时间戳
	if record.CreatedTs != nil {
		recordTime := time.Unix(*record.CreatedTs, 0)
		content.WriteString(fmt.Sprintf("**%s**\n\n", recordTime.Format("15:04:05")))
	}

	// 添加内容
	if record.Content != nil {
		content.WriteString(*record.Content)
		content.WriteString("\n\n")
	}

	// 添加资源列表
	if len(record.ResourceList) > 0 {
		content.WriteString("**附件：**\n")
		for _, resource := range record.ResourceList {
			if resource.Filename != nil {
				content.WriteString(fmt.Sprintf("- %s", *resource.Filename))
				if resource.ExternalLink != nil {
					content.WriteString(fmt.Sprintf(" ([链接](%s))", *resource.ExternalLink))
				}
				content.WriteString("\n")
			}
		}
		content.WriteString("\n")
	}

	content.WriteString("---\n\n")
	return content.String()
}

// getSummaryFrontMatter 获取摘要的Front Matter
func (e *MemosExporter) getSummaryFrontMatter() string {
	frontMatter := map[string]interface{}{
		"updated_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	content := "---\n"
	for key, value := range frontMatter {
		if value != nil {
			content += fmt.Sprintf("%s: %v\n", key, value)
		}
	}
	content += "---\n\n"
	return content
}
