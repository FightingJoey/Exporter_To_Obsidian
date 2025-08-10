package main

import (
	"fmt"
	"log"
	"io/fs"
	"os"
	"strconv"
	"time"
	"path/filepath"
	"strings"

	"exporter-to-obsidian/internal/client"
	"exporter-to-obsidian/internal/exporter"
	"exporter-to-obsidian/internal/types"
	"exporter-to-obsidian/internal/utils"

	"github.com/joho/godotenv"
)

// preprocessTasks 预处理任务时间字段
func preprocessTasks(tasks []types.Task) {
	for i := range tasks {
		if tasks[i].StartDate != nil {
			if parsed := utils.ParseDateTime(*tasks[i].StartDate); parsed != nil {
				tasks[i].ProcessedStartDate = parsed
			}
		}
		if tasks[i].DueDate != nil {
			if parsed := utils.ParseDateTime(*tasks[i].DueDate); parsed != nil {
				// 如果是全天任务，将截止日期减去一天
				if tasks[i].IsAllDay != nil && *tasks[i].IsAllDay && *tasks[i].StartDate != *tasks[i].DueDate {
					adjusted := parsed.AddDate(0, 0, -1) // 减去一天
					tasks[i].ProcessedDueDate = &adjusted
				} else {
					tasks[i].ProcessedDueDate = parsed
				}
			}
		}
	}
}

// getProjectColumns 获取项目的 Columns，并将其导出为 Markdown 文档
func getProjectColumns(client *client.Dida365Client, projectID string) ([]types.Column, error) {
	// 获取项目列数据
	columns, err := client.GetProjectColumns(projectID)
	if err != nil {
		log.Printf("获取项目 %s 列信息失败: %v\n", projectID, err)
		return nil, err
	}

	// 创建输出目录
	// outputDir := utils.GetEnvOrDefault("OUTPUT_DIR", ".")
	// columnsDir := filepath.Join(outputDir, "Columns")
	// if err := os.MkdirAll(columnsDir, 0755); err != nil {
	// 	fmt.Printf("创建目录失败 %s: %v\n", columnsDir, err)
	// 	return columns, nil
	// }

	// // 生成Markdown文件
	// filename := fmt.Sprintf("%s-columns.md", projectID)
	// filepath := filepath.Join(columnsDir, filename)

	// // 构建Markdown内容
	// content := fmt.Sprintf("# 项目 %s 的列信息\n\n", projectID)
	// content += fmt.Sprintf("更新时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	// content += "| 列ID | 列名称 | 创建时间 | 修改时间 | 排序 |\n"
	// content += "|------|--------|----------|----------|------|\n"

	// for _, column := range columns {
	// 	createdTime := ""
	// 	if column.CreatedTime != nil {
	// 		createdTime = utils.FormatTime(*column.CreatedTime, "2006-01-02 15:04:05")
	// 	}

	// 	modifiedTime := ""
	// 	if column.ModifiedTime != nil {
	// 		modifiedTime = utils.FormatTime(*column.ModifiedTime, "2006-01-02 15:04:05")
	// 	}

	// 	sortOrder := ""
	// 	if column.SortOrder != nil {
	// 		sortOrder = fmt.Sprintf("%d", *column.SortOrder)
	// 	}

	// 	content += fmt.Sprintf("| %s | %s | %s | %s | %s |\n", 
	// 		column.ID, column.Name, createdTime, modifiedTime, sortOrder)
	// }

	// // 写入文件
	// if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
	// 	fmt.Printf("写入文件失败: %v\n", err)
	// 	return columns, nil
	// }

	// fmt.Printf("已导出 %d 个列信息到文件: %s\n", len(columns), filepath)
	return columns, nil
}

// getTasks 获取任务数据
func getTasks(client *client.Dida365Client) ([]types.Project, []types.Task, []types.Task, []types.Project, []types.Task, []types.Column, error) {
	log.Printf("正在获取滴答清单数据...")

	// 获取所有数据
	allData, err := client.GetAllData()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("获取所有数据失败: %v", err)
	}

	// 解析项目数据
	var projects []types.Project

	// 解析笔记项目数据
	var note_projects []types.Project

	// 解析待办任务数据
	var todoTasks []types.Task

	// 解析笔记任务数据
	var notes []types.Task

	var all_columns []types.Column

	if inboxID := client.GetInboxID(); inboxID != "" {
		inbox := types.Project{}
        inbox.ID = inboxID
        inbox.Name = "收集箱"
		projects = append(projects, inbox)
	}

	if projectsData, ok := allData["projectProfiles"].([]interface{}); ok {
		for _, p := range projectsData {
			if projectMap, ok := p.(map[string]interface{}); ok {
				project := types.Project{}
				if id, ok := projectMap["id"].(string); ok {
					project.ID = id
					columns, _ := getProjectColumns(client, id)
					if columns != nil {
						all_columns = append(all_columns, columns...)
						project.Columns = columns
					}
				}
				if name, ok := projectMap["name"].(string); ok {
					project.Name = name
				}
				if kind, ok := projectMap["kind"].(string); ok {
					project.Kind = &kind
				}
				// 添加其他字段的解析...
				if *project.Kind == "TASK" {
					projects = append(projects, project)
				} else if *project.Kind == "NOTE" {
					note_projects = append(note_projects, project)
				}
			}
		}
	}

	if syncTaskBean, ok := allData["syncTaskBean"].(map[string]interface{}); ok {
		if tasksData, ok := syncTaskBean["update"].([]interface{}); ok {
			for _, t := range tasksData {
				if taskMap, ok := t.(map[string]interface{}); ok {
					task := parseTaskFromMap(taskMap)
					if *task.Kind == "TEXT" || *task.Kind == "CHECKLIST" {
						todoTasks = append(todoTasks, task)
					} else if *task.Kind == "NOTE" {
						notes = append(notes, task)
					}
				}
			}
		}
	}

	// 获取已完成任务
	today := time.Now()
	// 计算当前月份的开始日期
	startDate := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	var endDate time.Time
	if today.Month() == time.December {
		endDate = time.Date(today.Year()+1, time.January, 1, 0, 0, 0, 0, today.Location()).Add(-time.Second)
	} else {
		endDate = time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, today.Location()).Add(-time.Second)
	}
	completedTasks, err := client.GetCompletedTasks(
		startDate.Format("2006-01-02 15:04:05"),
		endDate.Format("2006-01-02 15:04:05"),
		50,
	)
	if err != nil {
		log.Printf("获取已完成任务失败: %v\n", err)
		completedTasks = []types.Task{}
	}

	// 预处理任务时间字段
	preprocessTasks(todoTasks)
	preprocessTasks(completedTasks)

	log.Printf("获取到 %d 个项目，%d 个待办任务，%d 个已完成任务，%d 个笔记项目，%d 个笔记, %d 个分组\n",
		len(projects), len(todoTasks), len(completedTasks), len(note_projects), len(notes), len(all_columns))

	return projects, todoTasks, completedTasks, note_projects, notes, all_columns, nil
}

// parseTaskFromMap 从map解析任务
func parseTaskFromMap(taskMap map[string]interface{}) types.Task {
	task := types.Task{}

	if id, ok := taskMap["id"].(string); ok {
		task.ID = &id
	}
	if title, ok := taskMap["title"].(string); ok {
		task.Title = &title
	}
	if projectID, ok := taskMap["projectId"].(string); ok {
		task.ProjectID = &projectID
	}
	if startDate, ok := taskMap["startDate"].(string); ok {
		task.StartDate = &startDate
	}
	if dueDate, ok := taskMap["dueDate"].(string); ok {
		task.DueDate = &dueDate
	}
	if priority, ok := taskMap["priority"].(float64); ok {
		pri := int(priority)
		task.Priority = &pri
	}
	if status, ok := taskMap["status"].(float64); ok {
		stat := int(status)
		task.Status = &stat
	}
	if createdTime, ok := taskMap["createdTime"].(string); ok {
		task.CreatedTime = &createdTime
	}
	if modifiedTime, ok := taskMap["modifiedTime"].(string); ok {
		task.ModifiedTime = &modifiedTime
	}
	if completedTime, ok := taskMap["completedTime"].(string); ok {
		task.CompletedTime = &completedTime
	}
	if content, ok := taskMap["content"].(string); ok {
		task.Content = &content
	}
	if isAllDay, ok := taskMap["isAllDay"].(bool); ok {
		task.IsAllDay = &isAllDay
	}
	if kind, ok := taskMap["kind"].(string); ok {
		task.Kind = &kind
	}
	if columnID, ok := taskMap["columnId"].(string); ok {
		task.ColumnID = &columnID
	}

	// 解析任务项
	if items, ok := taskMap["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				taskItem := types.TaskItem{}
				if title, ok := itemMap["title"].(string); ok {
					taskItem.Title = &title
				}
				if completedTime, ok := itemMap["completedTime"].(string); ok {
					taskItem.CompletedTime = &completedTime
				}
				if status, ok := itemMap["status"].(int); ok {
					taskItem.Status = &status
				}
				task.Items = append(task.Items, taskItem)
			}
		}
	}

	// 解析子任务ID
	if childIds, ok := taskMap["childIds"].([]interface{}); ok {
		for _, childId := range childIds {
			if id, ok := childId.(string); ok {
				task.ChildIDs = append(task.ChildIDs, id)
			}
		}
	}

	// 解析父任务ID
	if parentId, ok := taskMap["parentId"].(string); ok {
		task.ParentID = &parentId
	}

	return task
}

// getHabits 获取习惯数据
func getHabits(client *client.Dida365Client) ([]types.Habit, *types.HabitCheckinsResponse, int, error) {
	log.Printf("正在获取习惯数据...")

	// 获取习惯列表
	habits_data, err := client.GetHabits()
	if err != nil {
		log.Printf("获取习惯列表失败: %v\n", err)
		return []types.Habit{}, nil, 0, nil
	}
	var habits = []types.Habit{}
	for _, habit := range habits_data {
		if *habit.Status == 0 {
			habits = append(habits, habit)
		}
	}

	// 获取习惯打卡记录
	todayStamp := utils.GetTodayStamp()

	if len(habits) == 0 {
		log.Printf("没有习惯打卡记录\n")
		return []types.Habit{}, &types.HabitCheckinsResponse{}, todayStamp, nil
	}

	afterStamp := strconv.Itoa(todayStamp)

	var habitIDs []string
	for _, habit := range habits {
		if habit.ID != nil {
			habitIDs = append(habitIDs, *habit.ID)
		}
	}

	checkins, err := client.GetHabitsCheckins(afterStamp, habitIDs)
	if err != nil {
		log.Printf("获取习惯打卡失败: %v\n", err)
		checkins = &types.HabitCheckinsResponse{}
	}

	log.Printf("获取到 %d 个习惯\n", len(habits))
	return habits, checkins, todayStamp, nil
}

// exportDida365 导出滴答清单数据
func exportDida365() error {
	// 创建滴答清单客户端
	client, err := client.NewDida365Client("", "")
	if err != nil {
		return fmt.Errorf("创建滴答清单客户端失败: %v", err)
	}

	// 获取任务数据
	projects, todoTasks, completedTasks, note_projects, notes, all_columns, err := getTasks(client)
	if err != nil {
		return err
	}

	// 获取习惯数据
	habits, checkins, todayStamp, err := getHabits(client)
	if err != nil {
		return err
	}

	// 创建导出器
	outputDir := utils.GetEnvOrDefault("OUTPUT_DIR", ".")
	exporter := exporter.NewDida365Exporter(projects, todoTasks, completedTasks, outputDir, note_projects, notes, all_columns)

	// 导出项目任务
	if err := exporter.ExportProjectTasks(); err != nil {
		return fmt.Errorf("导出项目任务失败: %v", err)
	}

	// 导出笔记
	if err := exporter.ExportNotes(); err != nil {
		return fmt.Errorf("导出笔记失败: %v", err)
	}

	// 导出分组
	if err := exporter.ExportColumns(); err != nil {
		return fmt.Errorf("导出分组失败: %v", err)
	}

	// 导出每日摘要
	today := time.Now()
	if err := exporter.ExportDailySummary(today, habits, checkins, todayStamp); err != nil {
		return fmt.Errorf("导出每日摘要失败: %v", err)
	}

	// 导出每周摘要
	if err := exporter.ExportWeeklySummary(today); err != nil {
		return fmt.Errorf("导出每周摘要失败: %v", err)
	}

	// 导出每月摘要
	if err := exporter.ExportMonthlySummary(today); err != nil {
		return fmt.Errorf("导出每月摘要失败: %v", err)
	}

	log.Printf("滴答清单数据导出完成")
	return nil
}

// exportMemos 导出Memos数据
func exportMemos() error {
	// 检查是否配置了Memos
	memosAPI := os.Getenv("MEMOS_API")
	memosToken := os.Getenv("MEMOS_TOKEN")
	if memosAPI == "" || memosToken == "" {
		log.Printf("未配置Memos API，跳过Memos导出")
		return nil
	}

	log.Printf("正在导出Memos数据...")

	// 创建Memos客户端
	client, err := client.NewMemosClient(memosAPI, memosToken)
	if err != nil {
		return fmt.Errorf("创建Memos客户端失败: %v", err)
	}

	// 获取Memos记录
	records, err := client.FetchMemos(10, 0, "NORMAL")
	if err != nil {
		return fmt.Errorf("获取Memos记录失败: %v", err)
	}

	log.Printf("获取到 %d 条Memos记录\n", len(records))

	// 创建导出器
	outputDir := utils.GetEnvOrDefault("OUTPUT_DIR", ".")
	exporter := exporter.NewMemosExporter(records, outputDir)

	// 导出每日摘要
	today := time.Now()
	if err := exporter.ExportDailyMemos(today); err != nil {
		return fmt.Errorf("导出Memos每日摘要失败: %v", err)
	}

	log.Printf("Memos数据导出完成")
	return nil
}

func removeConflictFiles() {
	searchPath := utils.GetEnvOrDefault("OUTPUT_DIR", ".")
	keyword := "sync-conflict"

	_ = filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".stversions" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.Contains(d.Name(), keyword) {
			_ = os.Remove(path)
		}
		return nil
	})

	log.Printf("已删除冲突文件")
}

// runExport 执行一次数据导出
func runExport() {
	log.Printf("开始导出数据...")

	// 导出滴答清单数据
	if err := exportDida365(); err != nil {
		log.Printf("导出滴答清单数据失败: %v", err)
	}

	// 导出Memos数据
	if err := exportMemos(); err != nil {
		log.Printf("导出Memos数据失败: %v", err)
	}

	log.Printf("数据导出完成")

	removeConflictFiles()
}

func main() {
	// 加载环境变量
	godotenv.Load()

	// 创建定时器，每5分钟触发一次
	ticker := time.NewTicker(5 * time.Minute)

	// 立即执行第一次导出
	runExport()

	// 进入无限循环，等待定时器触发
	for range ticker.C {
		runExport()
	}
}
