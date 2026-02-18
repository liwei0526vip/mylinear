package model

import (
	"reflect"
	"strings"
	"testing"
)

// Task 1.4: 编写 Issue GORM 模型扩展测试
// 验证 Position 字段和 Subscribers 关联

// TestIssueHasPositionField 测试 Issue 模型有 Position 字段
func TestIssueHasPositionField(t *testing.T) {
	issueType := reflect.TypeOf(Issue{})

	field, ok := issueType.FieldByName("Position")
	if !ok {
		t.Error("Issue 模型应该有 Position 字段")
		return
	}

	// 验证类型为 float64
	if field.Type.String() != "float64" {
		t.Errorf("Issue.Position 应该是 float64 类型，实际为: %s", field.Type.String())
	}

	// 验证 gorm 标签
	gormTag := field.Tag.Get("gorm")
	if !strings.Contains(gormTag, "not null") {
		t.Errorf("Issue.Position 的 gorm 标签应包含 not null，实际为: %s", gormTag)
	}
	if !strings.Contains(gormTag, "default:0") {
		t.Errorf("Issue.Position 的 gorm 标签应包含 default:0，实际为: %s", gormTag)
	}
}

// TestIssueHasSubscribersRelation 测试 Issue 模型有 Subscribers 关联
func TestIssueHasSubscribersRelation(t *testing.T) {
	issueType := reflect.TypeOf(Issue{})

	field, ok := issueType.FieldByName("Subscribers")
	if !ok {
		t.Error("Issue 模型应该有 Subscribers 字段")
		return
	}

	// 验证类型为切片
	if field.Type.Kind() != reflect.Slice {
		t.Errorf("Issue.Subscribers 应该是切片类型，实际为: %s", field.Type.Kind())
	}

	// 验证 gorm 标签包含外键约束
	gormTag := field.Tag.Get("gorm")
	if !strings.Contains(gormTag, "foreignKey") {
		t.Errorf("Issue.Subscribers 的 gorm 标签应包含 foreignKey，实际为: %s", gormTag)
	}
}

// Task 1.6: 编写 IssueSubscription GORM 模型测试
// 验证复合主键和关联关系

// TestIssueSubscriptionModelExists 测试 IssueSubscription 模型存在
func TestIssueSubscriptionModelExists(t *testing.T) {
	subType := reflect.TypeOf(IssueSubscription{})

	// 验证 IssueID 字段
	issueIDField, ok := subType.FieldByName("IssueID")
	if !ok {
		t.Error("IssueSubscription 模型应该有 IssueID 字段")
		return
	}
	if issueIDField.Type.String() != "uuid.UUID" {
		t.Errorf("IssueSubscription.IssueID 应该是 uuid.UUID 类型，实际为: %s", issueIDField.Type.String())
	}

	// 验证 UserID 字段
	userIDField, ok := subType.FieldByName("UserID")
	if !ok {
		t.Error("IssueSubscription 模型应该有 UserID 字段")
		return
	}
	if userIDField.Type.String() != "uuid.UUID" {
		t.Errorf("IssueSubscription.UserID 应该是 uuid.UUID 类型，实际为: %s", userIDField.Type.String())
	}

	// 验证 CreatedAt 字段
	createdAtField, ok := subType.FieldByName("CreatedAt")
	if !ok {
		t.Error("IssueSubscription 模型应该有 CreatedAt 字段")
		return
	}
	if createdAtField.Type.String() != "time.Time" {
		t.Errorf("IssueSubscription.CreatedAt 应该是 time.Time 类型，实际为: %s", createdAtField.Type.String())
	}
}

// TestIssueSubscriptionRelations 测试 IssueSubscription 关联关系
func TestIssueSubscriptionRelations(t *testing.T) {
	subType := reflect.TypeOf(IssueSubscription{})

	// 验证 Issue 关联
	issueField, ok := subType.FieldByName("Issue")
	if !ok {
		t.Error("IssueSubscription 模型应该有 Issue 关联字段")
		return
	}

	gormTag := issueField.Tag.Get("gorm")
	if !strings.Contains(gormTag, "foreignKey:IssueID") {
		t.Errorf("IssueSubscription.Issue 的 gorm 标签应包含 foreignKey:IssueID，实际为: %s", gormTag)
	}

	// 验证 User 关联
	userField, ok := subType.FieldByName("User")
	if !ok {
		t.Error("IssueSubscription 模型应该有 User 关联字段")
		return
	}

	gormTag = userField.Tag.Get("gorm")
	if !strings.Contains(gormTag, "foreignKey:UserID") {
		t.Errorf("IssueSubscription.User 的 gorm 标签应包含 foreignKey:UserID，实际为: %s", gormTag)
	}
}

// TestIssueSubscriptionTableName 测试 IssueSubscription 表名
func TestIssueSubscriptionTableName(t *testing.T) {
	sub := IssueSubscription{}
	expected := "issue_subscriptions"
	if sub.TableName() != expected {
		t.Errorf("IssueSubscription.TableName() 应该返回 %s，实际为: %s", expected, sub.TableName())
	}
}

// TestIssueIdentifier 测试 Issue 标识符生成
func TestIssueIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		issue    Issue
		teamKey  string
		expected string
	}{
		{
			name:     "带团队 Key 的标识符",
			issue:    Issue{Number: 123},
			teamKey:  "ENG",
			expected: "ENG-123",
		},
		{
			name:     "不带团队 Key 的标识符",
			issue:    Issue{Number: 456},
			teamKey:  "",
			expected: "456",
		},
		{
			name:     "大数字",
			issue:    Issue{Number: 99999},
			teamKey:  "DEV",
			expected: "DEV-99999",
		},
		{
			name:     "小数字",
			issue:    Issue{Number: 1},
			teamKey:  "TEST",
			expected: "TEST-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.issue.Identifier(tt.teamKey)
			if result != tt.expected {
				t.Errorf("Identifier(%q) = %q, 期望 %q", tt.teamKey, result, tt.expected)
			}
		})
	}
}
