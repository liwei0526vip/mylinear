package service

import (
	"testing"
)

// TestParseMentions 测试 @mention 解析
func TestParseMentions(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "单个 @mention",
			body:     "这是给 @alice 的评论",
			expected: []string{"alice"},
		},
		{
			name:     "多个 @mention",
			body:     "@alice 和 @bob 请看一下",
			expected: []string{"alice", "bob"},
		},
		{
			name:     "无 @mention",
			body:     "这是普通评论，没有提及",
			expected: []string{},
		},
		{
			name:     "邮箱地址不匹配",
			body:     "联系 user@example.com 或 @alice",
			expected: []string{"alice"},
		},
		{
			name:     "带下划线的用户名",
			body:     "@user_name 和 @another_user",
			expected: []string{"user_name", "another_user"},
		},
		{
			name:     "带数字的用户名",
			body:     "@user123 和 @alice456",
			expected: []string{"user123", "alice456"},
		},
		{
			name:     "连续的 @mention（无空格分隔，只匹配第一个）",
			body:     "@alice@bob@charlie",
			expected: []string{"alice"}, // 无空格分隔时，后续 @ 不匹配
		},
		{
			name:     "空格分隔的连续 @mention",
			body:     "@alice @bob @charlie",
			expected: []string{"alice", "bob", "charlie"},
		},
		{
			name:     "@ 在行首",
			body:     "@alice\n这是新的一行",
			expected: []string{"alice"},
		},
		{
			name:     "中文用户名后的 @",
			body:     "你好@alice_world",
			expected: []string{"alice_world"},
		},
		{
			name:     "空字符串",
			body:     "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMentions(tt.body)
			if len(got) != len(tt.expected) {
				t.Errorf("ParseMentions() got %v, want %v", got, tt.expected)
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("ParseMentions()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestParseMentions_Deduplication 测试去重
func TestParseMentions_Deduplication(t *testing.T) {
	body := "@alice 提到了 @alice 多次 @alice"
	got := ParseMentions(body)

	// 统计 alice 出现次数
	counts := make(map[string]int)
	for _, name := range got {
		counts[name]++
	}

	if counts["alice"] != 3 {
		// 注意：ParseMentions 不做去重，返回所有匹配
		t.Errorf("ParseMentions() 期望返回 3 个 alice，实际返回 %d 个", counts["alice"])
	}
}

// TestExtractUniqueMentions 测试提取唯一用户名
func TestExtractUniqueMentions(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "重复的 @mention 去重",
			body:     "@alice 和 @alice 再次 @alice",
			expected: []string{"alice"},
		},
		{
			name:     "多个不同用户",
			body:     "@alice @bob @alice @charlie",
			expected: []string{"alice", "bob", "charlie"},
		},
		{
			name:     "空字符串",
			body:     "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractUniqueMentions(tt.body)
			if len(got) != len(tt.expected) {
				t.Errorf("ExtractUniqueMentions() got %v, want %v", got, tt.expected)
				return
			}
			// 检查所有期望的值都在结果中
			gotMap := make(map[string]bool)
			for _, v := range got {
				gotMap[v] = true
			}
			for _, v := range tt.expected {
				if !gotMap[v] {
					t.Errorf("ExtractUniqueMentions() 缺少 %v", v)
				}
			}
		})
	}
}
