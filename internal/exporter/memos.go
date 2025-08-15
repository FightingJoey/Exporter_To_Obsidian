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
	records   []types.MemosRecord
	outputDir string
	memosDir  string
}

// NewMemosExporter 创建新的Memos导出器
func NewMemosExporter(records []types.MemosRecord, outputDir string) *MemosExporter {
	if outputDir == "" {
		outputDir = utils.GetEnvOrDefault("OUTPUT_DIR", ".")
	}

	memosDir := filepath.Join(outputDir, utils.GetEnvOrDefault("MEMOS_DIR", "Memos"))

	exporter := &MemosExporter{
		records:   records,
		outputDir: outputDir,
		memosDir:  memosDir,
	}

	// 确保目录存在
	dirs := []string{
		exporter.memosDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", dir, err)
		}
	}

	return exporter
}

// ExportDailyMemos 导出每日Memos摘要
func (e *MemosExporter) ExportDailyMemos(date time.Time) error {
	// 设置日期范围
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endDate := startDate.Add(24*time.Hour - time.Second)

	// 获取当日的Memos记录
	dailyRecords := e.getRecordsInDateRange(startDate, endDate)

	// 创建文件名
	filename := fmt.Sprintf("%s-Memos.md", date.Format("2006-01-02"))
	filepath := filepath.Join(e.memosDir, filename)

	// 准备文件内容
	content := utils.GetFrontMatter([]string{"noyaml"}, "")
	// content += fmt.Sprintf("# %s Memos摘要\n\n", date.Format("2006-01-02"))

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
		return fmt.Errorf("今日没有Memos记录")
	}

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入每日Memos摘要失败: %v", err)
	}

	fmt.Printf("已创建每日Memos摘要：%s\n", filename)
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

// formatMemosRecord 格式化Memos记录
func (e *MemosExporter) formatMemosRecord(record types.MemosRecord) string {
	var content strings.Builder

	// 添加时间戳
	if record.CreatedTs != nil {
		recordTime := time.Unix(*record.CreatedTs, 0)
		content.WriteString(fmt.Sprintf("- **%s**\n\n", recordTime.Format("15:04:05")))
	}

	// 添加内容
	if record.Content != nil {
		content.WriteString(fmt.Sprintf("\t%s", *record.Content))
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
