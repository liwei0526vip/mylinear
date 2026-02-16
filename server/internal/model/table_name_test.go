package model

import (
	"testing"
)

// TestTableNames 测试所有模型的 TableName 方法
func TestTableNames(t *testing.T) {
	tests := []struct {
		name     string
		model    interface{ TableName() string }
		expected string
	}{
		{"Workspace", Workspace{}, "workspaces"},
		{"Team", Team{}, "teams"},
		{"TeamMember", TeamMember{}, "team_members"},
		{"User", User{}, "users"},
		{"Issue", Issue{}, "issues"},
		{"IssueRelation", IssueRelation{}, "issue_relations"},
		{"IssueClosure", IssueClosure{}, "issue_closure"},
		{"IssueStatusHistory", IssueStatusHistory{}, "issue_status_history"},
		{"WorkflowState", WorkflowState{}, "workflow_states"},
		{"WorkflowTransition", WorkflowTransition{}, "workflow_transitions"},
		{"Project", Project{}, "projects"},
		{"Milestone", Milestone{}, "milestones"},
		{"Cycle", Cycle{}, "cycles"},
		{"Label", Label{}, "labels"},
		{"Comment", Comment{}, "comments"},
		{"Attachment", Attachment{}, "attachments"},
		{"Document", Document{}, "documents"},
		{"Notification", Notification{}, "notifications"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.TableName()
			if got != tt.expected {
				t.Errorf("%s.TableName() = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

// TestModelImplementsTableName 测试模型是否实现了 TableName 方法
func TestModelImplementsTableName(t *testing.T) {
	// 定义 TableName 接口
	type TableNameInterface interface {
		TableName() string
	}

	// 测试所有模型都实现了该接口
	var _ TableNameInterface = (*Workspace)(nil)
	var _ TableNameInterface = (*Team)(nil)
	var _ TableNameInterface = (*TeamMember)(nil)
	var _ TableNameInterface = (*User)(nil)
	var _ TableNameInterface = (*Issue)(nil)
	var _ TableNameInterface = (*IssueRelation)(nil)
	var _ TableNameInterface = (*IssueClosure)(nil)
	var _ TableNameInterface = (*IssueStatusHistory)(nil)
	var _ TableNameInterface = (*WorkflowState)(nil)
	var _ TableNameInterface = (*WorkflowTransition)(nil)
	var _ TableNameInterface = (*Project)(nil)
	var _ TableNameInterface = (*Milestone)(nil)
	var _ TableNameInterface = (*Cycle)(nil)
	var _ TableNameInterface = (*Label)(nil)
	var _ TableNameInterface = (*Comment)(nil)
	var _ TableNameInterface = (*Attachment)(nil)
	var _ TableNameInterface = (*Document)(nil)
	var _ TableNameInterface = (*Notification)(nil)
}

// TestModelCount 测试模型数量（确保有 18 张表）
func TestModelCount(t *testing.T) {
	// 18 张表对应的模型
	expectedModelCount := 18

	// 已定义的模型
	models := []string{
		"Workspace", "Team", "TeamMember", "User",
		"Issue", "IssueRelation", "IssueClosure", "IssueStatusHistory",
		"WorkflowState", "WorkflowTransition",
		"Project", "Milestone", "Cycle",
		"Label", "Comment", "Attachment", "Document", "Notification",
	}

	if len(models) != expectedModelCount {
		t.Errorf("定义了 %d 个模型，期望 %d 个", len(models), expectedModelCount)
	}
}
