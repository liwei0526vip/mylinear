// Package service 提供业务逻辑层
package service

import (
	"regexp"
)

// mentionRegex 匹配 @username 格式的正则表达式
// 用户名规则：字母、数字、下划线
// 使用 (?:^|[^\w]) 确保前面是非单词字符或行首，排除邮箱地址中的 @
var mentionRegex = regexp.MustCompile(`(?:^|[^\w])@([a-zA-Z0-9_]+)`)

// ParseMentions 从文本中解析 @mention，返回所有匹配的用户名
// 注意：返回的列表可能包含重复项
func ParseMentions(body string) []string {
	if body == "" {
		return []string{}
	}

	matches := mentionRegex.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return []string{}
	}

	usernames := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			usernames = append(usernames, match[1])
		}
	}

	return usernames
}

// ExtractUniqueMentions 从文本中提取唯一的 @mention 用户名
func ExtractUniqueMentions(body string) []string {
	usernames := ParseMentions(body)
	if len(usernames) == 0 {
		return []string{}
	}

	// 去重
	seen := make(map[string]bool)
	unique := make([]string, 0, len(usernames))
	for _, name := range usernames {
		if !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}

	return unique
}
