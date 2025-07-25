# Exporter_To_Obsidian

## 项目描述

这是一个Go语言编写的工具，用于将滴答清单（Dida365）的数据导出到Obsidian兼容的Markdown格式。同时支持导出Memos数据。项目包括任务导出、习惯打卡、每日/每周/每月摘要生成等功能。

## 主要功能

- 滴答清单导出 ：
  - 获取项目、任务（待办和已完成）、习惯数据。
  - 导出项目任务到Markdown文件。
  - 生成每日、每周、每月摘要，包括任务和习惯打卡。
- Memos导出 ：
  - 获取Memos记录。
  - 生成每日摘要。
- 支持Docker部署 ：通过Dockerfile和docker-compose.yml实现容器化部署和定时任务。

## 安装

### 环境要求

- Go 1.x（详见go.mod）
- Docker（可选，用于容器化）

### 步骤

1. 克隆仓库：

   ```bash
   git clone <repository-url>
   ```
2. 安装依赖：
   
   ```bash
   go mod tidy
   ```

3. 编译：
   
   ```bash
   go build -o main ./cmd/main.go
   ```

## 配置

- 创建 .env 文件（参考 env.example ）：
  - DIDA365_USERNAME ：滴答清单用户名
  - DIDA365_PASSWORD ：滴答清单密码
  - MEMOS_API ：Memos API URL
  - MEMOS_TOKEN ：Memos访问令牌
  - OUTPUT_DIR ：输出目录（默认当前目录）
  - CALENDAR_DIR ：日历目录（默认"Calendar"）
  - TASKS_DIR ：任务目录（默认"Tasks"）
  - TASKS_INBOX_PATH ：任务收件箱路径（默认"Inbox"）

## 使用

### Docker部署

1. 构建镜像：
   
   ```bash
   docker build -t exporter-to-obsidian .
   ```

2. 启动容器：
   
   ```bash
   docker-compose up -d
   ```

- 支持定时任务（通过 docker_crontab 配置）。

## 项目结构

- cmd/main.go ：程序入口
- internal/client/ ：API客户端（Dida365和Memos）
- internal/exporter/ ：数据导出逻辑
- internal/types/ ：数据类型定义
- internal/utils/ ：工具函数
- Dockerfile ：Docker镜像定义
- docker-compose.yml ：容器编排

## 许可证

详见LICENSE文件。