package model

import (
	"reflect"
	"strings"
	"testing"
)

// TestModelHasUUIDPrimaryKey 测试所有模型都有 UUID 主键
// 注意：TeamMember 和 IssueClosure 使用复合主键，不在此测试中
func TestModelHasUUIDPrimaryKey(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
	}{
		{"Workspace", reflect.TypeOf(Workspace{})},
		{"Team", reflect.TypeOf(Team{})},
		{"User", reflect.TypeOf(User{})},
		{"Issue", reflect.TypeOf(Issue{})},
		{"IssueRelation", reflect.TypeOf(IssueRelation{})},
		{"IssueStatusHistory", reflect.TypeOf(IssueStatusHistory{})},
		{"WorkflowState", reflect.TypeOf(WorkflowState{})},
		{"WorkflowTransition", reflect.TypeOf(WorkflowTransition{})},
		{"Project", reflect.TypeOf(Project{})},
		{"Milestone", reflect.TypeOf(Milestone{})},
		{"Cycle", reflect.TypeOf(Cycle{})},
		{"Label", reflect.TypeOf(Label{})},
		{"Comment", reflect.TypeOf(Comment{})},
		{"Attachment", reflect.TypeOf(Attachment{})},
		{"Document", reflect.TypeOf(Document{})},
		{"Notification", reflect.TypeOf(Notification{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 检查是否有 ID 字段
			idField, ok := tt.modelType.FieldByName("ID")
			if !ok {
				// 检查是否嵌入 Model（有 Model 字段）
				if modelField, ok := tt.modelType.FieldByName("Model"); ok {
					// 嵌入的 Model 结构体的字段
					modelType := modelField.Type
					idField, ok = modelType.FieldByName("ID")
					if !ok {
						t.Errorf("%s 嵌入的 Model 结构体没有 ID 字段", tt.name)
						return
					}
				} else if modelField, ok := tt.modelType.FieldByName("ModelWithSoftDelete"); ok {
					// 嵌入的 ModelWithSoftDelete 结构体
					modelType := modelField.Type
					idField, ok = modelType.FieldByName("ID")
					if !ok {
						t.Errorf("%s 嵌入的 ModelWithSoftDelete 结构体没有 ID 字段", tt.name)
						return
					}
				} else {
					t.Errorf("%s 没有 ID 字段", tt.name)
					return
				}
			}

			// 检查 ID 字段的 gorm 标签包含 uuid 类型
			gormTag := idField.Tag.Get("gorm")
			if !strings.Contains(gormTag, "type:uuid") && !strings.Contains(gormTag, "uuid") {
				t.Errorf("%s.ID 的 gorm 标签应包含 type:uuid，实际为: %s", tt.name, gormTag)
			}

			// 检查 ID 字段是否为主键
			if !strings.Contains(gormTag, "primary_key") && !strings.Contains(gormTag, "primaryKey") {
				t.Errorf("%s.ID 的 gorm 标签应包含 primary_key，实际为: %s", tt.name, gormTag)
			}
		})
	}
}

// TestModelHasTimestamps 测试模型有时间戳字段
func TestModelHasTimestamps(t *testing.T) {
	tests := []struct {
		name      string
		modelType reflect.Type
	}{
		{"Workspace", reflect.TypeOf(Workspace{})},
		{"Team", reflect.TypeOf(Team{})},
		{"User", reflect.TypeOf(User{})},
		{"Issue", reflect.TypeOf(Issue{})},
		{"Comment", reflect.TypeOf(Comment{})},
		{"Project", reflect.TypeOf(Project{})},
		{"Document", reflect.TypeOf(Document{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 检查 CreatedAt 字段
			createdAtField, ok := tt.modelType.FieldByName("CreatedAt")
			if !ok {
				// 检查嵌入字段
				if modelField, ok := tt.modelType.FieldByName("Model"); ok {
					createdAtField, ok = modelField.Type.FieldByName("CreatedAt")
				} else if modelField, ok := tt.modelType.FieldByName("ModelWithSoftDelete"); ok {
					createdAtField, ok = modelField.Type.FieldByName("CreatedAt")
				}
			}

			if !ok {
				t.Errorf("%s 没有 CreatedAt 字段", tt.name)
			} else if createdAtField.Type.String() != "time.Time" {
				t.Errorf("%s.CreatedAt 应该是 time.Time 类型，实际为: %s", tt.name, createdAtField.Type.String())
			}

			// 检查 UpdatedAt 字段
			updatedAtField, ok := tt.modelType.FieldByName("UpdatedAt")
			if !ok {
				// 检查嵌入字段
				if modelField, ok := tt.modelType.FieldByName("Model"); ok {
					updatedAtField, ok = modelField.Type.FieldByName("UpdatedAt")
				} else if modelField, ok := tt.modelType.FieldByName("ModelWithSoftDelete"); ok {
					updatedAtField, ok = modelField.Type.FieldByName("UpdatedAt")
				}
			}

			if !ok {
				t.Errorf("%s 没有 UpdatedAt 字段", tt.name)
			} else if updatedAtField.Type.String() != "time.Time" {
				t.Errorf("%s.UpdatedAt 应该是 time.Time 类型，实际为: %s", tt.name, updatedAtField.Type.String())
			}
		})
	}
}

// TestForeignKeyFields 测试外键字段类型
func TestForeignKeyFields(t *testing.T) {
	tests := []struct {
		name         string
		modelType    reflect.Type
		fkField      string
		expectUUID   bool
	}{
		{"Team.WorkspaceID", reflect.TypeOf(Team{}), "WorkspaceID", true},
		{"Team.ParentID", reflect.TypeOf(Team{}), "ParentID", true},
		{"User.WorkspaceID", reflect.TypeOf(User{}), "WorkspaceID", true},
		{"Issue.TeamID", reflect.TypeOf(Issue{}), "TeamID", true},
		{"Issue.StatusID", reflect.TypeOf(Issue{}), "StatusID", true},
		{"Issue.AssigneeID", reflect.TypeOf(Issue{}), "AssigneeID", true},
		{"Issue.ProjectID", reflect.TypeOf(Issue{}), "ProjectID", true},
		{"Comment.IssueID", reflect.TypeOf(Comment{}), "IssueID", true},
		{"Comment.UserID", reflect.TypeOf(Comment{}), "UserID", true},
		{"Notification.UserID", reflect.TypeOf(Notification{}), "UserID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := tt.modelType.FieldByName(tt.fkField)
			if !ok {
				t.Errorf("%s 没有 %s 字段", tt.name, tt.fkField)
				return
			}

			// 检查外键类型
			fieldType := field.Type.String()
			isUUID := strings.Contains(fieldType, "uuid.UUID")

			if tt.expectUUID && !isUUID {
				// 检查是否为指针类型的 UUID
				if !strings.Contains(fieldType, "*uuid.UUID") && !strings.Contains(fieldType, "uuid.UUID") {
					t.Errorf("%s.%s 应该是 uuid.UUID 类型，实际为: %s", tt.name, tt.fkField, fieldType)
				}
			}
		})
	}
}

// TestEnumFieldTypes 测试枚举字段类型
func TestEnumFieldTypes(t *testing.T) {
	tests := []struct {
		name       string
		modelType  reflect.Type
		enumField  string
		expectType string
	}{
		{"User.Role", reflect.TypeOf(User{}), "Role", "model.Role"},
		{"TeamMember.Role", reflect.TypeOf(TeamMember{}), "Role", "model.Role"},
		{"WorkflowState.Type", reflect.TypeOf(WorkflowState{}), "Type", "model.StateType"},
		{"IssueRelation.Type", reflect.TypeOf(IssueRelation{}), "Type", "model.IssueRelationType"},
		{"Project.Status", reflect.TypeOf(Project{}), "Status", "model.ProjectStatus"},
		{"Cycle.Status", reflect.TypeOf(Cycle{}), "Status", "model.CycleStatus"},
		{"Notification.Type", reflect.TypeOf(Notification{}), "Type", "model.NotificationType"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := tt.modelType.FieldByName(tt.enumField)
			if !ok {
				t.Errorf("%s 没有 %s 字段", tt.name, tt.enumField)
				return
			}

			if field.Type.String() != tt.expectType {
				t.Errorf("%s.%s 应该是 %s 类型，实际为: %s", tt.name, tt.enumField, tt.expectType, field.Type.String())
			}
		})
	}
}

// TestArrayFieldTypes 测试数组字段类型
func TestArrayFieldTypes(t *testing.T) {
	tests := []struct {
		name       string
		modelType  reflect.Type
		arrayField string
	}{
		{"Issue.Labels", reflect.TypeOf(Issue{}), "Labels"},
		{"Project.Teams", reflect.TypeOf(Project{}), "Teams"},
		{"Project.Labels", reflect.TypeOf(Project{}), "Labels"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := tt.modelType.FieldByName(tt.arrayField)
			if !ok {
				t.Errorf("%s 没有 %s 字段", tt.name, tt.arrayField)
				return
			}

			// 检查是否为切片类型
			if field.Type.Kind() != reflect.Slice {
				t.Errorf("%s.%s 应该是切片类型，实际为: %s", tt.name, tt.arrayField, field.Type.Kind())
			}
		})
	}
}
