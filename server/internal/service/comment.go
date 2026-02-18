// Package service 提供业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/store"
)

// CommentService 定义评论业务逻辑接口
type CommentService interface {
	// CreateComment 创建评论
	CreateComment(ctx context.Context, issueID, userID uuid.UUID, body string, parentID *uuid.UUID) (*model.Comment, error)
	// UpdateComment 更新评论（仅作者可更新）
	UpdateComment(ctx context.Context, commentID, userID uuid.UUID, body string) (*model.Comment, error)
	// DeleteComment 删除评论（仅作者或管理员可删除）
	DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error
	// GetCommentsByIssueID 获取 Issue 的评论列表
	GetCommentsByIssueID(ctx context.Context, issueID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error)
}

// commentService 实现 CommentService 接口
type commentService struct {
	commentStore      store.CommentStore
	issueStore        store.IssueStore
	subscriptionStore store.IssueSubscriptionStore
	userStore         store.UserStore
}

// NewCommentService 创建评论服务实例
func NewCommentService(
	commentStore store.CommentStore,
	issueStore store.IssueStore,
	subscriptionStore store.IssueSubscriptionStore,
	userStore store.UserStore,
) CommentService {
	return &commentService{
		commentStore:      commentStore,
		issueStore:        issueStore,
		subscriptionStore: subscriptionStore,
		userStore:         userStore,
	}
}

// CreateComment 创建评论
func (s *commentService) CreateComment(ctx context.Context, issueID, userID uuid.UUID, body string, parentID *uuid.UUID) (*model.Comment, error) {
	// 创建评论
	comment := &model.Comment{
		IssueID:  issueID,
		UserID:   userID,
		Body:     body,
		ParentID: parentID,
	}

	if err := s.commentStore.CreateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	// 自动订阅评论者
	if err := s.subscriptionStore.Subscribe(ctx, issueID, userID); err != nil {
		// 订阅失败不影响评论创建，记录日志即可
		// TODO: 添加日志记录
	}

	// 解析 @mentions 并自动订阅
	usernames := ExtractUniqueMentions(body)
	for _, username := range usernames {
		// 查找被提及的用户
		user, err := s.userStore.GetUserByUsername(ctx, username)
		if err != nil {
			continue // 用户不存在，跳过
		}
		// 自动订阅被提及的用户
		_ = s.subscriptionStore.Subscribe(ctx, issueID, user.ID)
	}

	return comment, nil
}

// UpdateComment 更新评论（仅作者可更新）
func (s *commentService) UpdateComment(ctx context.Context, commentID, userID uuid.UUID, body string) (*model.Comment, error) {
	// 获取评论
	comment, err := s.commentStore.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// 权限校验：仅作者可编辑
	if comment.UserID != userID {
		return nil, fmt.Errorf("无权限编辑此评论")
	}

	// 更新评论
	if err := s.commentStore.UpdateComment(ctx, commentID, body); err != nil {
		return nil, fmt.Errorf("更新评论失败: %w", err)
	}

	// 重新获取更新后的评论
	return s.commentStore.GetCommentByID(ctx, commentID)
}

// DeleteComment 删除评论（仅作者或管理员可删除）
func (s *commentService) DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error {
	// 获取评论
	comment, err := s.commentStore.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("获取评论失败: %w", err)
	}

	// 获取用户信息检查是否为管理员
	user, err := s.userStore.GetUserByID(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 权限校验：仅作者或管理员可删除
	if comment.UserID != userID && user.Role != model.RoleAdmin && user.Role != model.RoleGlobalAdmin {
		return fmt.Errorf("无权限删除此评论")
	}

	// 删除评论
	if err := s.commentStore.DeleteComment(ctx, commentID); err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}

	return nil
}

// GetCommentsByIssueID 获取 Issue 的评论列表
func (s *commentService) GetCommentsByIssueID(ctx context.Context, issueID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	opts := &store.ListCommentsOptions{
		Page:     page,
		PageSize: pageSize,
	}

	return s.commentStore.GetCommentsByIssueIDWithTotal(ctx, issueID, opts)
}
