package model

import (
	"database/sql/driver"
	"testing"
)

// TestRole 测试 Role 枚举类型
func TestRole(t *testing.T) {
	tests := []struct {
		name      string
		role      Role
		wantValid bool
		scanValue interface{}
		wantValue driver.Value
	}{
		{
			name:      "全局管理员有效",
			role:      RoleGlobalAdmin,
			wantValid: true,
			scanValue: "global_admin",
			wantValue: "global_admin",
		},
		{
			name:      "团队管理员有效",
			role:      RoleAdmin,
			wantValid: true,
			scanValue: "admin",
			wantValue: "admin",
		},
		{
			name:      "普通成员有效",
			role:      RoleMember,
			wantValid: true,
			scanValue: "member",
			wantValue: "member",
		},
		{
			name:      "访客有效",
			role:      RoleGuest,
			wantValid: true,
			scanValue: "guest",
			wantValue: "guest",
		},
		{
			name:      "无效角色",
			role:      Role("invalid"),
			wantValid: false,
			scanValue: "invalid",
			wantValue: "invalid",
		},
		{
			name:      "空角色",
			role:      Role(""),
			wantValid: false,
			scanValue: nil,
			wantValue: "",
		},
		{
			name:      "字节切片扫描",
			role:      RoleGlobalAdmin,
			wantValid: true,
			scanValue: []byte("global_admin"),
			wantValue: "global_admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 Valid 方法
			if got := tt.role.Valid(); got != tt.wantValid {
				t.Errorf("Role.Valid() = %v, want %v", got, tt.wantValid)
			}

			// 测试 Value 方法
			gotValue, err := tt.role.Value()
			if err != nil {
				t.Errorf("Role.Value() error = %v", err)
			}
			if gotValue != tt.wantValue {
				t.Errorf("Role.Value() = %v, want %v", gotValue, tt.wantValue)
			}

			// 测试 Scan 方法
			var scannedRole Role
			if err := scannedRole.Scan(tt.scanValue); err != nil {
				t.Errorf("Role.Scan() error = %v", err)
			}
			if tt.scanValue != nil && scannedRole != tt.role {
				t.Errorf("Role.Scan() = %v, want %v", scannedRole, tt.role)
			}
		})
	}
}

// TestStateType 测试 StateType 枚举类型
func TestStateType(t *testing.T) {
	tests := []struct {
		name      string
		stateType StateType
		wantValid bool
	}{
		{"待办有效", StateTypeBacklog, true},
		{"未开始有效", StateTypeUnstarted, true},
		{"进行中有效", StateTypeStarted, true},
		{"已完成有效", StateTypeCompleted, true},
		{"已取消有效", StateTypeCanceled, true},
		{"无效类型", StateType("invalid"), false},
		{"空类型", StateType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stateType.Valid(); got != tt.wantValid {
				t.Errorf("StateType.Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestIssueRelationType 测试 IssueRelationType 枚举类型
func TestIssueRelationType(t *testing.T) {
	tests := []struct {
		name      string
		relation  IssueRelationType
		wantValid bool
	}{
		{"被阻塞有效", RelationBlockedBy, true},
		{"阻塞有效", RelationBlocking, true},
		{"相关有效", RelationRelated, true},
		{"重复有效", RelationDuplicate, true},
		{"无效类型", IssueRelationType("invalid"), false},
		{"空类型", IssueRelationType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.relation.Valid(); got != tt.wantValid {
				t.Errorf("IssueRelationType.Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestProjectStatus 测试 ProjectStatus 枚举类型
func TestProjectStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    ProjectStatus
		wantValid bool
	}{
		{"计划中有效", ProjectStatusPlanned, true},
		{"进行中有效", ProjectStatusInProgress, true},
		{"已暂停有效", ProjectStatusPaused, true},
		{"已完成有效", ProjectStatusCompleted, true},
		{"已取消有效", ProjectStatusCancelled, true},
		{"无效类型", ProjectStatus("invalid"), false},
		{"空类型", ProjectStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.wantValid {
				t.Errorf("ProjectStatus.Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestCycleStatus 测试 CycleStatus 枚举类型
func TestCycleStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    CycleStatus
		wantValid bool
	}{
		{"即将开始有效", CycleStatusUpcoming, true},
		{"进行中有效", CycleStatusActive, true},
		{"已完成有效", CycleStatusCompleted, true},
		{"无效类型", CycleStatus("invalid"), false},
		{"空类型", CycleStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.wantValid {
				t.Errorf("CycleStatus.Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestNotificationType 测试 NotificationType 枚举类型
func TestNotificationType(t *testing.T) {
	tests := []struct {
		name       string
		notifyType NotificationType
		wantValid  bool
	}{
		{"Issue 分配有效", NotificationTypeIssueAssigned, true},
		{"Issue 提及有效", NotificationTypeIssueMentioned, true},
		{"Issue 评论有效", NotificationTypeIssueCommented, true},
		{"Issue 状态变更有效", NotificationTypeIssueStatusChanged, true},
		{"项目更新有效", NotificationTypeProjectUpdated, true},
		{"迭代开始有效", NotificationTypeCycleStarted, true},
		{"迭代结束有效", NotificationTypeCycleEnded, true},
		{"无效类型", NotificationType("invalid"), false},
		{"空类型", NotificationType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notifyType.Valid(); got != tt.wantValid {
				t.Errorf("NotificationType.Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestPriorityIsValid 测试优先级验证函数
func TestPriorityIsValid(t *testing.T) {
	tests := []struct {
		name      string
		priority  int
		wantValid bool
	}{
		{"无优先级有效", PriorityNone, true},
		{"紧急有效", PriorityUrgent, true},
		{"高优先级有效", PriorityHigh, true},
		{"中优先级有效", PriorityMedium, true},
		{"低优先级有效", PriorityLow, true},
		{"负数无效", -1, false},
		{"超出范围无效", 5, false},
		{"大数无效", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PriorityIsValid(tt.priority); got != tt.wantValid {
				t.Errorf("PriorityIsValid(%d) = %v, want %v", tt.priority, got, tt.wantValid)
			}
		})
	}
}

// TestEnumScanErrors 测试枚举类型的错误扫描
func TestEnumScanErrors(t *testing.T) {
	tests := []struct {
		name      string
		scanFn    func(interface{}) error
		scanValue interface{}
		wantErr   bool
	}{
		{
			name: "Role 扫描无效类型",
			scanFn: func(v interface{}) error {
				var r Role
				return r.Scan(v)
			},
			scanValue: 123,
			wantErr:   true,
		},
		{
			name: "StateType 扫描无效类型",
			scanFn: func(v interface{}) error {
				var s StateType
				return s.Scan(v)
			},
			scanValue: 123,
			wantErr:   true,
		},
		{
			name: "IssueRelationType 扫描无效类型",
			scanFn: func(v interface{}) error {
				var t IssueRelationType
				return t.Scan(v)
			},
			scanValue: 123,
			wantErr:   true,
		},
		{
			name: "ProjectStatus 扫描无效类型",
			scanFn: func(v interface{}) error {
				var p ProjectStatus
				return p.Scan(v)
			},
			scanValue: 123,
			wantErr:   true,
		},
		{
			name: "CycleStatus 扫描无效类型",
			scanFn: func(v interface{}) error {
				var c CycleStatus
				return c.Scan(v)
			},
			scanValue: 123,
			wantErr:   true,
		},
		{
			name: "NotificationType 扫描无效类型",
			scanFn: func(v interface{}) error {
				var n NotificationType
				return n.Scan(v)
			},
			scanValue: 123,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scanFn(tt.scanValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
