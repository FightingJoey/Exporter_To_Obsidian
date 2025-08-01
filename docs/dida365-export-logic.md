# 滴答清单任务导出逻辑

本文档详细说明了滴答清单(Dida365)任务导出到Obsidian的逻辑和流程。

## 整体架构

滴答清单导出功能主要由以下几个组件构成：

1. **Dida365Client** - 滴答清单API客户端，负责与滴答清单服务器通信
2. **Dida365Exporter** - 滴答清单导出器，负责将数据转换为Markdown格式并保存
3. **Types** - 数据类型定义，包括任务、项目、习惯等结构
4. **Utils** - 工具函数，提供时间处理、环境变量读取等辅助功能

## 数据获取流程

### 1. 用户认证
- 通过环境变量 `DIDA365_USERNAME` 和 `DIDA365_PASSWORD` 获取用户凭证
- 如果存在有效的 `DIDA365_TOKEN` 且未过期(24小时内)，则直接使用
- 否则通过 [/user/signon](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/client/dida365.go#L141-L154) 接口登录获取新token

### 2. 获取所有数据
- 调用 [/batch/check/0](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/client/dida365.go#L227-L244) 接口获取所有数据，包括：
  - 项目列表(projectProfiles)
  - 任务列表(syncTaskBean.update)
  - 其他元数据

### 3. 获取已完成任务
- 调用 [/project/all/completed](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/client/dida365.go#L247-L271) 接口获取已完成任务
- 默认获取当月的已完成任务

### 4. 获取习惯数据
- 调用 [/habits](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/client/dida365.go#L274-L292) 接口获取习惯列表
- 调用 [/habitCheckins/query](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/client/dida365.go#L295-L316) 接口获取习惯打卡记录

## 数据处理逻辑

### 时间处理
- 对于全天任务，截止日期会减去一天以正确反映任务时间范围
- 支持多种时间格式解析，包括ISO格式、带时区格式等
- 所有时间统一转换为东八区(北京时间)处理

### 任务关联处理
- 解析父子任务关系，建立任务间的层级结构
- 解析任务列表项，支持子任务项的完成状态
- 处理任务与项目的关联关系

## 导出文件结构

导出的文件按照以下结构组织：

```
输出目录/
├── Calendar/
│   ├── 1.Daily/
│   ├── 2.Weekly/
│   └── 3.Monthly/
├── Tasks/
└── Inbox/
```

### 1. 任务文件导出
- 每个任务导出为单独的Markdown文件，文件名为任务ID
- 文件保存在 [Tasks](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/exporter/dida365.go#L39-L39) 目录下
- 包含Front Matter元数据和任务详细信息
- 支持跳过未更新的任务文件以提高效率

### 2. 项目索引导出
- 所有项目任务汇总到 [TasksInbox.md](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/exporter/dida365.go#L46-L46) 文件中
- 按项目分组显示任务列表
- 任务按优先级排序

### 3. 日常摘要导出
- 每日摘要文件保存在 [Calendar/1.Daily](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/exporter/dida365.go#L41-L41) 目录下
- 文件名格式为 `YYYY-MM-DD-Dida365.md`
- 包含当日习惯打卡情况和任务完成情况

### 4. 每周摘要导出
- 每周摘要文件保存在 [Calendar/2.Weekly](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/exporter/dida365.go#L42-L42) 目录下
- 文件名格式为 `YYYY-WXX-Dida365.md`
- 按天展示一周内的任务安排

### 5. 每月摘要导出
- 每月摘要文件保存在 [Calendar/3.Monthly](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/exporter/dida365.go#L43-L43) 目录下
- 文件名格式为 `YYYY-MM-Dida365.md`
- 按周展示一个月内的任务安排

## 特殊功能

### 图片URL转换
- 自动识别并转换滴答清单中的图片引用格式
- 将 `![image](<attachment_id>/<filename>)` 转换为完整的URL格式

### 超链接转换
- 自动识别并转换滴答清单中的超链接格式
- 将 `[链接文本](url)` 转换为 Markdown 格式 `[taskId|链接文本]`

### 优先级标记
- 将数字优先级转换为可视化标记：
  - 0(默认): ⏬
  - 1(低): 🔽
  - 3(高): 🔼
  - 5(紧急): ⏫

## 环境变量配置

主要环境变量包括：
- `DIDA365_USERNAME`: 滴答清单用户名
- `DIDA365_PASSWORD`: 滴答清单密码
- `DIDA365_TOKEN`: 登录token(自动生成和维护)
- `OUTPUT_DIR`: 输出目录路径
- `CALENDAR_DIR`: 日历目录名称(默认为Calendar)
- `TASKS_DIR`: 任务目录名称(默认为Tasks)
- `TASKS_INBOX_PATH`: 任务收件箱目录名称(默认为Inbox)