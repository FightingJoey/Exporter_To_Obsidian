package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	"regexp"

	"exporter-to-obsidian/internal/types"
	"exporter-to-obsidian/internal/utils"
)

// Dida365Exporter 滴答清单导出器
type Dida365Exporter struct {
	projects       []types.Project
	todoTasks      []types.Task
	completedTasks []types.Task
	note_projects  []types.Project
	notes          []types.Task
	all_columns    []types.Column
	outputDir      string
	calendarDir    string
	dailyDir       string
	weeklyDir      string
	monthlyDir     string
	tasksDir       string
	tasksInboxDir  string
	tasksInboxPath string
	notesDir       string
	columnsDir     string
}

// NewDida365Exporter 创建新的滴答清单导出器
func NewDida365Exporter(projects []types.Project, todoTasks, completedTasks []types.Task, outputDir string, note_projects []types.Project, notes []types.Task, all_columns []types.Column) *Dida365Exporter {
	if outputDir == "" {
		outputDir = os.Getenv("OUTPUT_DIR")
		if outputDir == "" {
			wd, _ := os.Getwd()
			outputDir = wd
		}
	}

	calendarDir := filepath.Join(outputDir, utils.GetEnvOrDefault("CALENDAR_DIR", "Calendar"))
	tasksDir := filepath.Join(outputDir, utils.GetEnvOrDefault("TASKS_DIR", "Tasks"))
	tasksInboxDir := filepath.Join(outputDir, utils.GetEnvOrDefault("TASKS_INBOX_PATH", "Inbox"))
	notesDir := filepath.Join(outputDir, utils.GetEnvOrDefault("NOTES_DIR", "Notes"))
	columnsDir := filepath.Join(outputDir, utils.GetEnvOrDefault("COLUMNS_DIR", "Columns"))

	exporter := &Dida365Exporter{
		projects:       projects,
		todoTasks:      todoTasks,
		completedTasks: completedTasks,
		note_projects:  note_projects,
		notes:          notes,
		all_columns:    all_columns,
		outputDir:      outputDir,
		calendarDir:    calendarDir,
		dailyDir:       filepath.Join(calendarDir, "1.Daily"),
		weeklyDir:      filepath.Join(calendarDir, "2.Weekly"),
		monthlyDir:     filepath.Join(calendarDir, "3.Monthly"),
		tasksDir:       tasksDir,
		tasksInboxDir:  tasksInboxDir,
		tasksInboxPath: filepath.Join(tasksInboxDir, "TasksInbox.md"),
		notesDir:       notesDir,
		columnsDir:     columnsDir,
	}

	// 确保所有目录存在
	dirs := []string{
		exporter.calendarDir,
		exporter.dailyDir,
		exporter.weeklyDir,
		exporter.monthlyDir,
		exporter.tasksDir,
		exporter.tasksInboxDir,
		exporter.notesDir,
		exporter.columnsDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", dir, err)
		}
	}

	return exporter
}

func (e *Dida365Exporter) ExportNotes() error {
	// 为每个笔记创建Markdown文件
	for _, note := range e.notes {
		if err := e.createNoteMarkdown(note); err != nil {
			fmt.Printf("创建笔记文件失败: %v\n", err)
			return fmt.Errorf("创建笔记文件失败: %v", err)
		}
	}
	return nil
}

func (e *Dida365Exporter) createNoteMarkdown(task types.Task) error {
	if task.ID == nil {
		return fmt.Errorf("笔记ID为空")
	}

	filename := fmt.Sprintf("%s.md", *task.ID)
	filepath := filepath.Join(e.notesDir, filename)

	// 检查文件是否需要更新
	if e.shouldSkipFile(filepath, task) {
		// fmt.Printf("笔记文件已是最新: %s\n", filename)
		return nil
	}

	// 准备Front Matter
	content := e.buildTaskFrontMatter(task)

	// 添加任务描述
	if task.Content != nil && *task.Content != "" {
		convertedContent := e.convertImageURLs(*task.Content, *task.ProjectID, *task.ID)
		content += fmt.Sprintf("%s\n\n", convertedContent)
	}

	// 删除旧文件并写入新文件
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
		// fmt.Printf("删除旧文件: %s\n", filename)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入笔记文件失败: %v", err)
	}
	return nil
}

func (e *Dida365Exporter) ExportColumns() error {
	// 为每个分组创建Markdown文件
	for _, column := range e.all_columns {
		if err := e.createColumnMarkdown(column); err != nil {
			fmt.Printf("创建分组文件失败: %v\n", err)
			return fmt.Errorf("创建分组文件失败: %v", err)
		}
	}
	return nil
}

func (e *Dida365Exporter) createColumnMarkdown(column types.Column) error {
	if column.ID == nil {
		return fmt.Errorf("分组ID为空")
	}

	filename := fmt.Sprintf("%s.md", *column.ID)
	filepath := filepath.Join(e.columnsDir, filename)

	if e.skip(filepath) {
		// fmt.Printf("文件已存在: %s\n", filename)
		return nil
	}

	frontMatter := ""
	write := func(k string, v interface{}) {
		if v != nil {
			frontMatter += fmt.Sprintf("%s: %v\n", k, v)
		}
	}
	write("title", *column.Name)
	write("column_id", *column.ID)
	write("project_id", *column.ProjectID)
	write("created_time", utils.FormatTime(*column.CreatedTime, "2006-01-02 15:04:05"))
	write("modified_time", utils.FormatTime(*column.ModifiedTime, "2006-01-02 15:04:05"))
	content := utils.GetFrontMatter([]string{"fullwidth", "noyaml"}, frontMatter)

	for _, project := range e.projects {
		if *column.ProjectID == project.ID {
			content += fmt.Sprintf("# %s —— %s\n\n", project.Name, *column.Name)
			break
		}
	}

	content += "```dataviewjs\n" +
		"dv.view('dida365TaskTable', {\n" +
		"    folderPath: '9.Archive/Dida365/Tasks',\n" +
		"    condition: (p, c) => {\n" +
		"        return p.frontmatter?.column_id === c.frontmatter?.column_id;\n" +
		"    }\n" +
		"});\n" +
		"```\n\n"

	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
		// fmt.Printf("删除旧文件: %s\n", filename)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入分组文件失败: %v", err)
	}

	return nil
}

// ExportProjectTasks 导出所有项目的任务
func (e *Dida365Exporter) ExportProjectTasks() error {
	// 构建任务映射
	allTasks := append(e.todoTasks, e.completedTasks...)
	taskMap := make(map[string]types.Task)
	for _, task := range allTasks {
		if task.ID != nil {
			taskMap[*task.ID] = task
		}
	}

	// 创建项目索引内容
	allContent := utils.GetFrontMatter([]string{"noyaml"}, "")

	// 为每个项目生成内容
	for _, project := range e.projects {
		projectTasks := e.getProjectTasks(project.ID, e.todoTasks)
		// 为每个任务创建Markdown文件
		for _, task := range projectTasks {
			if err := e.createTaskMarkdown(task, taskMap); err != nil {
				fmt.Printf("创建任务文件失败: %v\n", err)
			}
		}
		allContent += e.getProjectIndexContent(project, projectTasks)
	}

	// 为已完成任务创建Markdown文件
	for _, task := range e.completedTasks {
		if err := e.createTaskMarkdown(task, taskMap); err != nil {
			fmt.Printf("创建已完成任务文件失败: %v\n", err)
		}
	}

	// 写入项目索引文件
	if err := os.WriteFile(e.tasksInboxPath, []byte(allContent), 0644); err != nil {
		return fmt.Errorf("写入项目索引文件失败: %v", err)
	}

	fmt.Println("已创建统一项目索引文件: TasksInbox.md")
	return nil
}

// getProjectTasks 获取指定项目的任务
func (e *Dida365Exporter) getProjectTasks(projectID string, tasks []types.Task) []types.Task {
	var projectTasks []types.Task
	for _, task := range tasks {
		if task.ProjectID != nil && *task.ProjectID == projectID {
			projectTasks = append(projectTasks, task)
		}
	}
	return projectTasks
}
func (e *Dida365Exporter) convertImageURLs(content, projectID, taskID string) string {
	// 正则匹配图片格式：![image](<attachment_id>/<filename>)
	re := regexp.MustCompile(`!\[image]\(([0-9a-f]+)/([^\)]+)\)`)
	
	// 替换为指定URL格式
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match // 不符合格式则返回原字符串
		}
		
		attachmentID := parts[1]
		newURL := fmt.Sprintf("https://dida365.com/api/v1/attachment/%s/%s/%s.jpg", 
			projectID, taskID, attachmentID)
		
		return fmt.Sprintf("![image](%s)", newURL)
	})

	// 转换任务链接格式
	content = e.convertTaskLinks(content)
	
	return content
}

// convertTaskLinks 将内容中的任务链接转换为内部链接格式
func (e *Dida365Exporter) convertTaskLinks(content string) string {
	// 正则匹配滴答清单任务链接格式：[链接文本](https://dida365.com/webapp/#p/{projectID}/tasks/{taskID})
	// 捕获组：1=链接文本, 2=projectID, 3=taskID
	re := regexp.MustCompile(`\[([^\]]+)\]\(https://dida365\.com/webapp/#p/([a-zA-Z0-9]+)/tasks/([a-zA-Z0-9]+)\)`)
	
	// 替换为Obsidian内部链接格式：[[taskID|链接文本]]
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match // 不符合格式则返回原字符串
		}
		
		linkText := parts[1]
		taskID := parts[3]
		
		return fmt.Sprintf("[[%s|%s]]", taskID, linkText)
	})
}

// createTaskMarkdown 为单个任务创建Markdown文件
func (e *Dida365Exporter) createTaskMarkdown(task types.Task, taskMap map[string]types.Task) error {
	if task.ID == nil {
		return fmt.Errorf("任务ID为空")
	}

	filename := fmt.Sprintf("%s.md", *task.ID)
	filepath := filepath.Join(e.tasksDir, filename)

	// 检查文件是否需要更新
	if e.shouldSkipFile(filepath, task) {
		// fmt.Printf("任务文件已是最新: %s\n", filename)
		return nil
	}

	// 准备Front Matter
	content := e.buildTaskFrontMatter(task)

	// 添加任务描述
	if task.Content != nil && *task.Content != "" {
		convertedContent := e.convertImageURLs(*task.Content, *task.ProjectID, *task.ID)
		content += fmt.Sprintf("%s\n\n", convertedContent)
	}

	if task.Desc != nil && *task.Desc != "" {
		convertedContent := e.convertImageURLs(*task.Desc, *task.ProjectID, *task.ID)
		content += fmt.Sprintf("%s\n\n", convertedContent)
	}

	// 添加任务列表
	if len(task.Items) > 0 {
		content += "## 任务列表\n\n"
		for _, item := range task.Items {
			status := " "
			if item.CompletedTime != nil {
				status = "x"
			}
			title := ""
			if item.Title != nil {
				title = *item.Title
			}
			content += fmt.Sprintf("- [%s] %s\n", status, title)
		}
		content += "\n"
	}

	// 添加子任务列表
	if len(task.ChildIDs) > 0 {
		content += "## 子任务列表\n\n"
		content += "```dataviewjs\n" +
			"dv.view('dida365TaskTable', {\n" +
			"    folderPath: '9.Archive/Dida365/Tasks',\n" +
			"    condition: (p, c) => {\n" +
			"        return p.frontmatter?.task_id === c.frontmatter?.parent_id;\n" +
			"    }\n" +
			"});\n" +
			"```\n\n"
		// content += e.createTableHeader()
		// for _, childID := range task.ChildIDs {
		// 	if childTask, exists := taskMap[childID]; exists {
		// 		content += e.createTaskTableContent(childTask)
		// 	} else {
		// 		content += fmt.Sprintf("| %s | %s | %s | %s | %s |\n", childID, "", "", "", "")
		// 	}
		// }
		// content += "\n"
	}

	// 添加父任务
	// if task.ParentID != nil && *task.ParentID != "" {
	// 	content += "## 父任务\n\n"
	// 	content += e.createTableHeader()
	// 	if parentTask, exists := taskMap[*task.ParentID]; exists {
	// 		content += e.createTaskTableContent(parentTask)
	// 	} else {
	// 		content += fmt.Sprintf("| %s | %s | %s | %s | %s |\n", *task.ParentID, "", "", "", "")
	// 	}
	// 	content += "\n"
	// }

	// 删除旧文件并写入新文件
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
		fmt.Printf("删除旧文件: %s\n", filename)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入任务文件失败: %v", err)
	}

	// fmt.Printf("已创建任务文件: %s\n", filename)
	return nil
}

// shouldSkipFile 检查是否应该跳过任务文件创建
func (e *Dida365Exporter) shouldSkipFile(filepath string, task types.Task) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return false
	}

	// 检查修改时间
	if task.ModifiedTime != nil {
		fileModifiedTime := utils.ExtractFrontMatterField(string(content), "modified_time")
		taskModifiedTime := utils.FormatTime(*task.ModifiedTime, "2006-01-02 15:04:05")
		return fileModifiedTime == taskModifiedTime
	}

	return false
}

// buildTaskFrontMatter 构建任务的Front Matter
func (e *Dida365Exporter) buildTaskFrontMatter(task types.Task) string {
	frontMatter := ""
	write := func(k string, v interface{}) {
		if v != nil {
			frontMatter += fmt.Sprintf("%s: %v\n", k, v)
		}
	}
	write("title", *task.Title)
	write("task_id", *task.ID)
	write("project_id", *task.ProjectID)
	write("column_id", *task.ColumnID)
	if task.ParentID != nil {
		write("parent_id", *task.ParentID)
	}
	write("priority", *task.Priority)
	write("status", *task.Status)
	if task.ProcessedStartDate != nil {
		write("start_date", task.ProcessedStartDate.Format("2006-01-02 15:04:05"))
	} else if task.StartDate != nil {
		write("start_date", utils.FormatTime(*task.StartDate, "2006-01-02 15:04:05"))
	}
	if task.ProcessedDueDate != nil {
		write("due_date", task.ProcessedDueDate.Format("2006-01-02 15:04:05"))
	} else if task.DueDate != nil {
		write("due_date", utils.FormatTime(*task.DueDate, "2006-01-02 15:04:05"))
	}
	write("created_time", utils.FormatTime(*task.CreatedTime, "2006-01-02 15:04:05"))
	write("modified_time", utils.FormatTime(*task.ModifiedTime, "2006-01-02 15:04:05"))
	if task.CompletedTime != nil {
		write("completed_time", utils.FormatTime(*task.CompletedTime, "2006-01-02 15:04:05"))
	}
	content := utils.GetFrontMatter([]string{"noyaml"}, frontMatter)
	content += fmt.Sprintf("# %s\n\n", *task.Title)
	return content
}

// getProjectIndexContent 获取项目索引内容
func (e *Dida365Exporter) getProjectIndexContent(project types.Project, tasks []types.Task) string {
	content := fmt.Sprintf("## %s\n\n", project.Name)

	if len(tasks) > 0 {
		// 创建列映射，将任务按列分组
		columnTasks := make(map[string][]types.Task)
		
		// 按列ID分组任务
		for _, task := range tasks {
			if task.ColumnID != nil && *task.ColumnID != "" {
				columnTasks[*task.ColumnID] = append(columnTasks[*task.ColumnID], task)
			}
		}
		
		// 创建列ID到列名称的映射
		columnNameMap := make(map[string]string)
		for _, column := range project.Columns {
			columnNameMap[*column.ID] = *column.Name
		}
		
		// 按列显示任务
		// 首先显示有列的任务
		for _, column := range project.Columns {
			tasksInColumn := columnTasks[*column.ID]
			if len(tasksInColumn) > 0 {
				// 显示列标题
				content += fmt.Sprintf("### %s\n\n", *column.Name)
				
				// 对列内的任务按优先级排序
				sort.Slice(tasksInColumn, func(i, j int) bool {
					priI := 0
					if tasksInColumn[i].Priority != nil {
						priI = *tasksInColumn[i].Priority
					}
					priJ := 0
					if tasksInColumn[j].Priority != nil {
						priJ = *tasksInColumn[j].Priority
					}
					if priI != priJ {
						return priI > priJ
					}
					// 如果优先级相同，按创建时间排序
					createdI := ""
					if tasksInColumn[i].CreatedTime != nil {
						createdI = *tasksInColumn[i].CreatedTime
					}
					createdJ := ""
					if tasksInColumn[j].CreatedTime != nil {
						createdJ = *tasksInColumn[j].CreatedTime
					}
					return createdI < createdJ
				})
				
				// 显示任务列表
				for _, task := range tasksInColumn {
					priorityMark := utils.GetPriorityMark(task.Priority)
					timeRange := e.formatTaskTimeRange(task)
					title := ""
					if task.Title != nil {
						title = *task.Title
					}
					id := ""
					if task.ID != nil {
						id = *task.ID
					}
					
					if timeRange == "" {
						content += fmt.Sprintf("- [ ] [[%s|%s]] | %s\n", id, title, priorityMark)
					} else {
						content += fmt.Sprintf("- [ ] [[%s|%s]] | %s | %s\n", id, title, priorityMark, timeRange)
					}
				}
				content += "\n"
			} else {
				content += fmt.Sprintf("### %s\n\n", *column.Name)
			}
		}
	}

	content += "\n"
	return content
}

// formatTaskTimeRange 格式化任务时间范围
func (e *Dida365Exporter) formatTaskTimeRange(task types.Task) string {
	var startDate, endDate string

	// 处理开始时间
	if task.ProcessedStartDate != nil {
		startDate = task.ProcessedStartDate.Format("2006-01-02")
	} else if task.StartDate != nil {
		startDate = utils.FormatTime(*task.StartDate, "2006-01-02")
	}

	// 处理结束时间
	if task.ProcessedDueDate != nil {
		endDate = task.ProcessedDueDate.Format("2006-01-02")
	} else if task.DueDate != nil {
		endDate = utils.FormatTime(*task.DueDate, "2006-01-02")
	}

	if startDate != "" && endDate != "" {
		if startDate == endDate {
			return fmt.Sprintf("📅 %s", endDate)
		}
		return fmt.Sprintf("🛫 %s ~ 📅 %s", startDate, endDate)
	} else if startDate != "" {
		return fmt.Sprintf("🛫 %s", startDate)
	} else if endDate != "" {
		return fmt.Sprintf("📅 %s", endDate)
	}

	return ""
}

// createTableHeader 创建表格头
func (e *Dida365Exporter) createTableHeader() string {
	return "| 任务 | 优先级 | 时间范围 | 状态 | 完成时间 |\n| --- | --- | --- | --- | --- |\n"
}

// createTaskTableContent 创建任务表格内容
func (e *Dida365Exporter) createTaskTableContent(task types.Task) string {
	title := ""
	if task.Title != nil {
		title = *task.Title
	}
	id := ""
	if task.ID != nil {
		id = *task.ID
	}

	titleLink := fmt.Sprintf("[[%s\\|%s]]", id, title)
	priorityMark := utils.GetPriorityMark(task.Priority)
	timeRange := e.formatTaskTimeRange(task)

	status := "待办"
	if task.Status != nil && *task.Status == 2 {
		status = "已完成"
	}

	doneTime := ""
	if task.Status != nil && *task.Status == 2 && task.CompletedTime != nil {
		doneTime = utils.FormatTime(*task.CompletedTime, "2006-01-02")
	}

	return fmt.Sprintf("| %s | %s | %s | %s | %s |\n", titleLink, priorityMark, timeRange, status, doneTime)
}

// ExportDailySummary 导出每日摘要
func (e *Dida365Exporter) ExportDailySummary(date time.Time, habits []types.Habit, checkins *types.HabitCheckinsResponse, todayStamp int) error {
	// 设置日期范围
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endDate := startDate.Add(24*time.Hour - time.Second)

	// 获取当日任务
	tasks := e.getTasksInDateRange(startDate, endDate)

	// 创建文件名
	filename := fmt.Sprintf("%s-Dida365.md", date.Format("2006-01-02"))
	filepath := filepath.Join(e.dailyDir, filename)

	// 准备文件内容
	content := utils.GetFrontMatter([]string{"noyaml"}, "")
	// content += fmt.Sprintf("\n# %s 摘要\n\n", date.Format("2006-01-02"))

	// 添加习惯打卡
	if len(habits) > 0 {
		content += "## 习惯打卡\n\n"
		for _, habit := range habits {
			checked := false
			doneDate := ""

			if checkins != nil && habit.ID != nil {
				if habitCheckins, exists := checkins.Checkins[*habit.ID]; exists {
					for _, checkin := range habitCheckins {
						if checkin.CheckinStamp != nil && *checkin.CheckinStamp == todayStamp &&
							checkin.Status != nil && *checkin.Status == 2 {
							checked = true
							if checkin.CheckinTime != nil {
								doneDate = utils.FormatTime(*checkin.CheckinTime, "2006-01-02")
							}
							break
						}
					}
				}
			}

			habitName := ""
			if habit.Name != nil {
				habitName = *habit.Name
			}

			if checked {
				content += fmt.Sprintf("- [x] %s | ✅ %s\n", habitName, doneDate)
			} else {
				content += fmt.Sprintf("- [ ] %s\n", habitName)
			}
		}
		content += "\n"
	}

	if len(tasks) > 0 {
		// 分离待办和已完成任务
		todoTasks := make([]types.Task, 0)
		doneTasks := make([]types.Task, 0)

		for _, task := range tasks {
			if task.Status != nil && *task.Status == 0 {
				todoTasks = append(todoTasks, task)
			} else if task.Status != nil && *task.Status == 2 {
				doneTasks = append(doneTasks, task)
			}
		}

		// 输出待办任务
		if len(todoTasks) > 0 {
			content += "## 待办任务\n\n"
			// 按优先级排序
			sort.Slice(todoTasks, func(i, j int) bool {
				priI := 0
				if todoTasks[i].Priority != nil {
					priI = *todoTasks[i].Priority
				}
				priJ := 0
				if todoTasks[j].Priority != nil {
					priJ = *todoTasks[j].Priority
				}
				return priI > priJ
			})

			for idx, task := range todoTasks {
				content += e.formatTaskLine(task, idx+1, true) + "\n"
			}
			content += "\n"
		}

		// 输出已完成任务
		if len(doneTasks) > 0 {
			content += "## 已完成任务\n\n"
			// 按优先级排序
			sort.Slice(doneTasks, func(i, j int) bool {
				priI := 0
				if doneTasks[i].Priority != nil {
					priI = *doneTasks[i].Priority
				}
				priJ := 0
				if doneTasks[j].Priority != nil {
					priJ = *doneTasks[j].Priority
				}
				return priI > priJ
			})

			for _, task := range doneTasks {
				content += e.formatTaskLine(task, 0, false) + "\n"
			}
			content += "\n"
		}
	} else {
		content += "今日没有任务。\n"
	}

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入每日摘要失败: %v", err)
	}

	fmt.Printf("已创建每日摘要：%s\n", filename)
	return nil
}

// getTasksInDateRange 获取指定日期范围内的任务
func (e *Dida365Exporter) getTasksInDateRange(startDate, endDate time.Time) []types.Task {
	var tasks []types.Task

	// 处理未完成任务
	for _, task := range e.todoTasks {
		if e.taskInRange(task, startDate, endDate) {
			tasks = append(tasks, task)
		}
	}

	// 处理已完成任务
	for _, task := range e.completedTasks {
		if e.taskInRange(task, startDate, endDate) {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// taskInRange 判断任务是否在指定时间范围内
func (e *Dida365Exporter) taskInRange(task types.Task, start, end time.Time) bool {
	var taskStart, taskEnd *time.Time

	// 获取任务开始时间
	if task.ProcessedStartDate != nil {
		taskStart = task.ProcessedStartDate
	} else if task.StartDate != nil {
		if parsed := utils.ParseDateTime(*task.StartDate); parsed != nil {
			taskStart = parsed
		}
	}

	// 获取任务结束时间
	if task.ProcessedDueDate != nil {
		taskEnd = task.ProcessedDueDate
	} else if task.DueDate != nil {
		if parsed := utils.ParseDateTime(*task.DueDate); parsed != nil {
			taskEnd = parsed
		}
	}

	if taskStart == nil && taskEnd == nil {
		return false
	}

	if taskStart != nil && taskEnd != nil {
		return !(taskEnd.Before(start) || taskStart.After(end))
	} else if taskStart != nil {
		return !taskStart.After(end)
	} else if taskEnd != nil {
		return !taskEnd.Before(start)
	}

	return false
}

// formatTaskLine 格式化任务行
func (e *Dida365Exporter) formatTaskLine(task types.Task, index int, ordered bool) string {
	priorityMark := utils.GetPriorityMark(task.Priority)
	timeRange := e.formatTaskTimeRange(task)

	var line string
	if ordered && index > 0 {
		title := ""
		if task.Title != nil {
			title = *task.Title
		}
		id := ""
		if task.ID != nil {
			id = *task.ID
		}
		line = fmt.Sprintf("%d. [[%s|%s]] | %s", index, id, title, priorityMark)
	} else {
		checkbox := " "
		if task.Status != nil && *task.Status == 2 {
			checkbox = "x"
		}
		title := ""
		if task.Title != nil {
			title = *task.Title
		}
		id := ""
		if task.ID != nil {
			id = *task.ID
		}
		line = fmt.Sprintf("- [%s] [[%s|%s]] | %s", checkbox, id, title, priorityMark)
	}

	if timeRange != "" {
		line += fmt.Sprintf(" | %s", timeRange)
	}

	// 对于已完成任务，添加 ✅ 和完成日期
	if task.Status != nil && *task.Status == 2 {
		doneDate := ""
		if task.CompletedTime != nil {
			doneDate = utils.FormatTime(*task.CompletedTime, "2006-01-02")
		}
		line += fmt.Sprintf(" | ✅ %s", doneDate)
	}

	return line
}

// ExportWeeklySummary 导出每周摘要
func (e *Dida365Exporter) ExportWeeklySummary(date time.Time) error {
	// 获取周的开始和结束日期（周一到周日）
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}
	startOfWeek := date.AddDate(0, 0, -(weekday - 1))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// 创建文件名
	year, week := time.Now().ISOWeek()
	filename := fmt.Sprintf("%d-W%d-Dida365.md", year, week)
	filepath := filepath.Join(e.weeklyDir, filename)

	if e.skip(filepath) {
		// fmt.Printf("文件已存在: %s\n", filename)
		return nil
	}

	// 准备文件内容
	content := utils.GetFrontMatter([]string{"fullwidth", "noyaml"}, "")
	content += fmt.Sprintf("# %d年第%02d周任务摘要\n\n", year, week)
	content += fmt.Sprintf("周期： %s 至 %s \n\n", startOfWeek.Format("2006-01-02"), endOfWeek.Format("2006-01-02"))
	content += e.dataviewjs(startOfWeek, endOfWeek)

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入每周摘要失败: %v", err)
	}

	fmt.Printf("已创建每周摘要：%s\n", filename)
	return nil
}

// ExportMonthlySummary 导出每月摘要
func (e *Dida365Exporter) ExportMonthlySummary(date time.Time) error {
	// 计算月份的第一天和最后一天
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	var lastDay time.Time
	if date.Month() == time.December {
		lastDay = time.Date(date.Year()+1, time.January, 1, 23, 59, 59, 0, time.Local).AddDate(0, 0, -1)
	} else {
		lastDay = time.Date(date.Year(), date.Month()+1, 1, 23, 59, 59, 0, time.Local).AddDate(0, 0, -1)
	}

	// 创建目录路径
	filename := fmt.Sprintf("%s-Dida365.md", date.Format("2006-01"))
	filepath := filepath.Join(e.monthlyDir, filename)

	if e.skip(filepath) {
		// fmt.Printf("文件已存在: %s\n", filename)
		return nil
	}

	// 构建Markdown内容
	content := utils.GetFrontMatter([]string{"fullwidth", "noyaml"}, "")
	content += fmt.Sprintf("# %s任务摘要\n\n", date.Format("2006年01月"))
	content += e.dataviewjs(firstDay, lastDay)

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入每月摘要失败: %v", err)
	}

	fmt.Printf("已创建每月摘要：%s\n", filename)
	return nil
}

func (e *Dida365Exporter) skip(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (e *Dida365Exporter) dataviewjs(start time.Time, end time.Time) string { 
	content := "```dataviewjs\n" +
            "dv.view('dida365TaskTable', {\n" +
            "    folderPath: '9.Archive/Dida365/Tasks',\n" +
            "    condition: (p, c) => {\n" +
			"        const rawDue = p.frontmatter?.due_date;\n" +
			"        const rawStart = p.frontmatter?.start_date;\n" +
            "        const normalize = v => {\n" +
			"            if (!v) return null;\n" +
			"            const str = String(v);\n" +
			"            const m = str.match(/^(\\d{4}-\\d{2}-\\d{2})/);\n" +
			"			 return m ? m[1] : null;\n" +
			"        };\n" +
			"        const dueStr = normalize(rawDue);\n" +
			"        const startStr = normalize(rawStart);\n" +
			"		 const dueDate   = dv.date(dueStr);\n" +
			"		 const startDate = dv.date(startStr);\n" +
			fmt.Sprintf("        const weekStart = dv.date('%s');\n", start.Format("2006-01-02")) +
			fmt.Sprintf("        const weekEnd   = dv.date('%s');\n", end.Format("2006-01-02")) +
			"        const inWeek = d => d && d >= weekStart && d <= weekEnd;\n" +
			"        return inWeek(dueDate) || inWeek(startDate);\n" +
            "    }\n" +
            "});\n" +
            "```\n\n"
	return content
}

// getTaskDate 获取任务所属的日期
func (e *Dida365Exporter) getTaskDate(task types.Task, startDate, endDate time.Time) string {
	var taskStart, taskEnd *time.Time

	// 获取任务开始时间
	if task.ProcessedStartDate != nil {
		taskStart = task.ProcessedStartDate
	} else if task.StartDate != nil {
		if parsed := utils.ParseDateTime(*task.StartDate); parsed != nil {
			taskStart = parsed
		}
	}

	// 获取任务结束时间
	if task.ProcessedDueDate != nil {
		taskEnd = task.ProcessedDueDate
	} else if task.DueDate != nil {
		if parsed := utils.ParseDateTime(*task.DueDate); parsed != nil {
			taskEnd = parsed
		}
	}

	// 如果有结束时间，使用结束时间的日期
	if taskEnd != nil && !taskEnd.Before(startDate) && !taskEnd.After(endDate) {
		return taskEnd.Format("2006-01-02")
	}

	// 如果有开始时间，使用开始时间的日期
	if taskStart != nil && !taskStart.Before(startDate) && !taskStart.After(endDate) {
		return taskStart.Format("2006-01-02")
	}

	return ""
}

// getTaskWeek 获取任务所属的周
func (e *Dida365Exporter) getTaskWeek(task types.Task, startOfMonth time.Time) string {
	var taskTime *time.Time

	// 优先使用结束时间，然后是开始时间
	if task.ProcessedDueDate != nil {
		taskTime = task.ProcessedDueDate
	} else if task.DueDate != nil {
		if parsed := utils.ParseDateTime(*task.DueDate); parsed != nil {
			taskTime = parsed
		}
	} else if task.ProcessedStartDate != nil {
		taskTime = task.ProcessedStartDate
	} else if task.StartDate != nil {
		if parsed := utils.ParseDateTime(*task.StartDate); parsed != nil {
			taskTime = parsed
		}
	}

	if taskTime == nil {
		return ""
	}

	// 计算是第几周
	weekday := int(taskTime.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := taskTime.AddDate(0, 0, -(weekday - 1))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// 计算周数
	weekNum := (startOfWeek.Day()-1)/7 + 1

	return fmt.Sprintf("第%d周 (%s ~ %s)", weekNum, startOfWeek.Format("01-02"), endOfWeek.Format("01-02"))
}
