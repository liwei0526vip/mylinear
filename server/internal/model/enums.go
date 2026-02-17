package model

import (
	"database/sql/driver"
	"fmt"
)

// Role 用户角色
type Role string

const (
	RoleGlobalAdmin Role = "global_admin" // 全局管理员
	RoleAdmin       Role = "admin"        // 团队管理员
	RoleMember      Role = "member"       // 普通成员
	RoleGuest       Role = "guest"        // 访客
)

// Valid 验证角色是否有效
func (r Role) Valid() bool {
	switch r {
	case RoleGlobalAdmin, RoleAdmin, RoleMember, RoleGuest:
		return true
	default:
		return false
	}
}

// Scan 实现 sql.Scanner 接口
func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*r = Role(v)
	case []byte:
		*r = Role(v)
	default:
		return fmt.Errorf("无法扫描 Role 类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

// StateType 工作流状态类型
type StateType string

const (
	StateTypeBacklog   StateType = "backlog"   // 待办
	StateTypeUnstarted StateType = "unstarted" // 未开始
	StateTypeStarted   StateType = "started"   // 进行中
	StateTypeCompleted StateType = "completed" // 已完成
	StateTypeCanceled  StateType = "canceled"  // 已取消
)

// Valid 验证状态类型是否有效
func (s StateType) Valid() bool {
	switch s {
	case StateTypeBacklog, StateTypeUnstarted, StateTypeStarted, StateTypeCompleted, StateTypeCanceled:
		return true
	default:
		return false
	}
}

// Scan 实现 sql.Scanner 接口
func (s *StateType) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = StateType(v)
	case []byte:
		*s = StateType(v)
	default:
		return fmt.Errorf("无法扫描 StateType 类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (s StateType) Value() (driver.Value, error) {
	return string(s), nil
}

// IssueRelationType Issue 关系类型
type IssueRelationType string

const (
	RelationBlockedBy IssueRelationType = "blocked_by" // 被阻塞
	RelationBlocking  IssueRelationType = "blocking"   // 阻塞
	RelationRelated   IssueRelationType = "related"    // 相关
	RelationDuplicate IssueRelationType = "duplicate"  // 重复
)

// Valid 验证关系类型是否有效
func (t IssueRelationType) Valid() bool {
	switch t {
	case RelationBlockedBy, RelationBlocking, RelationRelated, RelationDuplicate:
		return true
	default:
		return false
	}
}

// Scan 实现 sql.Scanner 接口
func (t *IssueRelationType) Scan(value interface{}) error {
	if value == nil {
		*t = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*t = IssueRelationType(v)
	case []byte:
		*t = IssueRelationType(v)
	default:
		return fmt.Errorf("无法扫描 IssueRelationType 类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (t IssueRelationType) Value() (driver.Value, error) {
	return string(t), nil
}

// ProjectStatus 项目状态
type ProjectStatus string

const (
	ProjectStatusPlanned    ProjectStatus = "planned"     // 计划中
	ProjectStatusInProgress ProjectStatus = "in_progress" // 进行中
	ProjectStatusPaused     ProjectStatus = "paused"      // 已暂停
	ProjectStatusCompleted  ProjectStatus = "completed"   // 已完成
	ProjectStatusCancelled  ProjectStatus = "cancelled"   // 已取消
)

// Valid 验证项目状态是否有效
func (p ProjectStatus) Valid() bool {
	switch p {
	case ProjectStatusPlanned, ProjectStatusInProgress, ProjectStatusPaused, ProjectStatusCompleted, ProjectStatusCancelled:
		return true
	default:
		return false
	}
}

// Scan 实现 sql.Scanner 接口
func (p *ProjectStatus) Scan(value interface{}) error {
	if value == nil {
		*p = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*p = ProjectStatus(v)
	case []byte:
		*p = ProjectStatus(v)
	default:
		return fmt.Errorf("无法扫描 ProjectStatus 类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (p ProjectStatus) Value() (driver.Value, error) {
	return string(p), nil
}

// CycleStatus 迭代状态
type CycleStatus string

const (
	CycleStatusUpcoming  CycleStatus = "upcoming"  // 即将开始
	CycleStatusActive    CycleStatus = "active"    // 进行中
	CycleStatusCompleted CycleStatus = "completed" // 已完成
)

// Valid 验证迭代状态是否有效
func (c CycleStatus) Valid() bool {
	switch c {
	case CycleStatusUpcoming, CycleStatusActive, CycleStatusCompleted:
		return true
	default:
		return false
	}
}

// Scan 实现 sql.Scanner 接口
func (c *CycleStatus) Scan(value interface{}) error {
	if value == nil {
		*c = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*c = CycleStatus(v)
	case []byte:
		*c = CycleStatus(v)
	default:
		return fmt.Errorf("无法扫描 CycleStatus 类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (c CycleStatus) Value() (driver.Value, error) {
	return string(c), nil
}

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeIssueAssigned      NotificationType = "issue_assigned"       // Issue 分配
	NotificationTypeIssueMentioned     NotificationType = "issue_mentioned"      // Issue 提及
	NotificationTypeIssueCommented     NotificationType = "issue_commented"      // Issue 评论
	NotificationTypeIssueStatusChanged NotificationType = "issue_status_changed" // Issue 状态变更
	NotificationTypeProjectUpdated     NotificationType = "project_updated"      // 项目更新
	NotificationTypeCycleStarted       NotificationType = "cycle_started"        // 迭代开始
	NotificationTypeCycleEnded         NotificationType = "cycle_ended"          // 迭代结束
)

// Valid 验证通知类型是否有效
func (n NotificationType) Valid() bool {
	switch n {
	case NotificationTypeIssueAssigned, NotificationTypeIssueMentioned,
		NotificationTypeIssueCommented, NotificationTypeIssueStatusChanged,
		NotificationTypeProjectUpdated, NotificationTypeCycleStarted,
		NotificationTypeCycleEnded:
		return true
	default:
		return false
	}
}

// Scan 实现 sql.Scanner 接口
func (n *NotificationType) Scan(value interface{}) error {
	if value == nil {
		*n = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*n = NotificationType(v)
	case []byte:
		*n = NotificationType(v)
	default:
		return fmt.Errorf("无法扫描 NotificationType 类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (n NotificationType) Value() (driver.Value, error) {
	return string(n), nil
}

// Issue 优先级常量
const (
	PriorityNone   = 0 // 无优先级
	PriorityUrgent = 1 // 紧急
	PriorityHigh   = 2 // 高
	PriorityMedium = 3 // 中
	PriorityLow    = 4 // 低
)

// PriorityIsValid 验证优先级是否有效
func PriorityIsValid(p int) bool {
	return p >= PriorityNone && p <= PriorityLow
}
