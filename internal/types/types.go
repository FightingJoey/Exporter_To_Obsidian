package types

import (
	"time"
)

// Tag 表示滴答清单中的一个标签
type Tag struct {
	Name      *string `json:"name,omitempty"`
	RawName   *string `json:"rawName,omitempty"`
	Label     *string `json:"label,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
	SortType  *string `json:"sortType,omitempty"`
	Color     *string `json:"color,omitempty"`
	Etag      *string `json:"etag,omitempty"`
	Type      *string `json:"type,omitempty"`
}

// Project 表示滴答清单中的一个项目（清单）
type Project struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	IsOwner              *bool   `json:"isOwner,omitempty"`
	Color                *string `json:"color,omitempty"`
	SortOrder            *int    `json:"sortOrder,omitempty"`
	SortOption           *string `json:"sortOption,omitempty"`
	SortType             *string `json:"sortType,omitempty"`
	UserCount            *int    `json:"userCount,omitempty"`
	Etag                 *string `json:"etag,omitempty"`
	ModifiedTime         *string `json:"modifiedTime,omitempty"`
	InAll                *bool   `json:"inAll,omitempty"`
	ShowType             *string `json:"showType,omitempty"`
	Muted                *bool   `json:"muted,omitempty"`
	ReminderType         *string `json:"reminderType,omitempty"`
	Closed               *bool   `json:"closed,omitempty"`
	Transferred          *bool   `json:"transferred,omitempty"`
	GroupID              *string `json:"groupId,omitempty"`
	ViewMode             *string `json:"viewMode,omitempty"`
	NotificationOptions  *string `json:"notificationOptions,omitempty"`
	TeamID               *string `json:"teamId,omitempty"`
	Permission           *string `json:"permission,omitempty"`
	Kind                 *string `json:"kind,omitempty"`
	Timeline             *string `json:"timeline,omitempty"`
	NeedAudit            *bool   `json:"needAudit,omitempty"`
	BarcodeNeedAudit     *bool   `json:"barcodeNeedAudit,omitempty"`
	OpenToTeam           *bool   `json:"openToTeam,omitempty"`
	TeamMemberPermission *string `json:"teamMemberPermission,omitempty"`
	Source               *string `json:"source,omitempty"`
}

// Task 表示滴答清单中的一个任务
type Task struct {
	ID            *string    `json:"id,omitempty"`
	Title         *string    `json:"title,omitempty"`
	ProjectID     *string    `json:"projectId,omitempty"`
	StartDate     *string    `json:"startDate,omitempty"`
	Items         []TaskItem `json:"items,omitempty"`
	ExDate        []string   `json:"exDate,omitempty"`
	DueDate       *string    `json:"dueDate,omitempty"`
	Priority      *int       `json:"priority,omitempty"`
	IsAllDay      *bool      `json:"isAllDay,omitempty"`
	RepeatFlag    *string    `json:"repeatFlag,omitempty"`
	Progress      *int       `json:"progress,omitempty"`
	Assignee      *string    `json:"assignee,omitempty"`
	SortOrder     *float64   `json:"sortOrder,omitempty"`
	IsFloating    *bool      `json:"isFloating,omitempty"`
	Status        *int       `json:"status,omitempty"`
	Kind          *string    `json:"kind,omitempty"`
	CreatedTime   *string    `json:"createdTime,omitempty"`
	ModifiedTime  *string    `json:"modifiedTime,omitempty"`
	CompletedTime *string    `json:"completedTime,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	TimeZone      *string    `json:"timeZone,omitempty"`
	Content       *string    `json:"content,omitempty"`
	Desc          *string    `json:"desc,omitempty"`
	ChildIDs      []string   `json:"childIds,omitempty"`
	ParentID      *string    `json:"parentId,omitempty"`
	// 预处理后的时间字段
	ProcessedStartDate *time.Time `json:"-"`
	ProcessedDueDate   *time.Time `json:"-"`
}

// TaskItem 表示任务中的子项
type TaskItem struct {
	Title         *string `json:"title,omitempty"`
	Status        *int    `json:"status,omitempty"`
	CompletedTime *string `json:"completedTime,omitempty"`
}

// Habit 表示滴答清单中的一个习惯
type Habit struct {
	ID              *string  `json:"id,omitempty"`
	Name            *string  `json:"name,omitempty"`
	IconRes         *string  `json:"iconRes,omitempty"`
	Color           *string  `json:"color,omitempty"`
	SortOrder       *int     `json:"sortOrder,omitempty"`
	Status          *int     `json:"status,omitempty"`
	Encouragement   *string  `json:"encouragement,omitempty"`
	TotalCheckIns   *int     `json:"totalCheckIns,omitempty"`
	CreatedTime     *string  `json:"createdTime,omitempty"`
	ModifiedTime    *string  `json:"modifiedTime,omitempty"`
	ArchivedTime    *string  `json:"archivedTime,omitempty"`
	Type            *string  `json:"type,omitempty"`
	Goal            *float64 `json:"goal,omitempty"` // 修改：从 *int 改为 *float64 以支持浮点数
	Step            *float64 `json:"step,omitempty"`
	Unit            *string  `json:"unit,omitempty"`
	Etag            *string  `json:"etag,omitempty"`
	RepeatRule      *string  `json:"repeatRule,omitempty"`
	RecordEnable    *bool    `json:"recordEnable,omitempty"`
	SectionID       *string  `json:"sectionId,omitempty"`
	TargetDays      *int     `json:"targetDays,omitempty"`
	TargetStartDate *int     `json:"targetStartDate,omitempty"`
	CompletedCycles *int     `json:"completedCycles,omitempty"`
	ExDates         []string `json:"exDates,omitempty"`
	Style           *int     `json:"style,omitempty"`
}

// MemosResource 表示Memos资源
type MemosResource struct {
	Name         *string `json:"name,omitempty"`
	ExternalLink *string `json:"externalLink,omitempty"`
	Type         *string `json:"type,omitempty"`
	UID          *string `json:"uid,omitempty"`
	ID           *int64  `json:"id,omitempty"` // 修改：从 *string 改为 *int64 以支持数字ID
	Filename     *string `json:"filename,omitempty"`
	Size         *int64  `json:"size,omitempty"`
}

// MemosRecord 表示Memos记录
type MemosRecord struct {
	RowStatus    *string         `json:"rowStatus,omitempty"`
	UpdatedTs    *int64          `json:"updatedTs,omitempty"`
	CreatedTs    *int64          `json:"createdTs,omitempty"`
	CreatedAt    *string         `json:"createdAt,omitempty"`
	UpdatedAt    *string         `json:"updatedAt,omitempty"`
	Content      *string         `json:"content,omitempty"`
	ResourceList []MemosResource `json:"resourceList,omitempty"`
}

// HabitCheckin 表示习惯打卡记录
type HabitCheckin struct {
	CheckinStamp *int    `json:"checkinStamp,omitempty"`
	Status       *int    `json:"status,omitempty"`
	CheckinTime  *string `json:"checkinTime,omitempty"`
}

// HabitCheckinsResponse 表示习惯打卡响应
type HabitCheckinsResponse struct {
	Checkins map[string][]HabitCheckin `json:"checkins,omitempty"`
}
