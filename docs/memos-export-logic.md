# Memos导出逻辑

本文档详细说明了Memos导出到Obsidian的逻辑和流程。

## 整体架构

Memos导出功能主要由以下几个组件构成：

1. **MemosClient** - Memos API客户端，负责与Memos服务器通信
2. **MemosExporter** - Memos导出器，负责将数据转换为Markdown格式并保存
3. **Types** - 数据类型定义，包括记录、资源等结构
4. **Utils** - 工具函数，提供时间处理、环境变量读取等辅助功能

## 数据获取流程

### 1. 用户认证
- 通过环境变量 `MEMOS_API` 和 `MEMOS_TOKEN` 获取API地址和访问令牌
- 在请求头中添加 `Authorization: Bearer <token>` 进行认证

### 2. 获取Memos记录
- 调用Memos API获取记录列表
- 默认参数为:
  - limit: 10 (获取记录数量)
  - offset: 0 (偏移量)
  - rowStatus: NORMAL (记录状态)

## 数据处理逻辑

### 时间处理
- Memos记录时间戳为Unix时间戳格式
- 所有时间统一转换为东八区(北京时间)处理
- 按日期对记录进行分组

### 资源处理
- 解析记录关联的资源(如图片、附件等)
- 提取资源文件名和外部链接

## 导出文件结构

导出的文件按照以下结构组织：

```
输出目录/
└── Memos/
```

### 日常摘要导出
- 每日摘要文件保存在 [Memos](file:///Users/joy/Desktop/Code/Exporter_To_Obsidian/internal/exporter/memos.go#L25-L25) 目录下
- 文件名格式为 `YYYY-MM-DD-Memos.md`
- 包含当日所有Memos记录，按时间倒序排列

## 导出内容格式

### 文件结构
- 每个文件包含Front Matter元数据
- 按时间顺序组织的Memos记录
- 每条记录包含时间戳、内容和附件信息

### 记录格式
每条Memos记录导出为以下格式：
```
**HH:MM:SS**

记录内容

**附件：**
- 文件名1 ([链接](外部链接))
- 文件名2 ([链接](外部链接))

---
```

## 环境变量配置

主要环境变量包括：
- `MEMOS_API`: Memos API地址
- `MEMOS_TOKEN`: Memos访问令牌
- `OUTPUT_DIR`: 输出目录路径
- `MEMOS_DIR`: Memos目录名称(默认为Memos)

## 导出逻辑细节

### 数据筛选
- 根据记录的创建时间筛选出当日的Memos记录
- 只导出状态为NORMAL的记录

### 排序规则
- 按创建时间倒序排列(最新的在前)

### 文件更新策略
- 每次运行都会重新生成当日的Memos摘要文件
- 不检查文件是否已存在或是否需要更新