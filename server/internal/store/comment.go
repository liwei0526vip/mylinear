// Package store 提供数据访问层
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/gorm"
)

// ListCommentsOptions 评论列表查询选项
type ListCommentsOptions struct {
	Page     int // 页码，从 1 开始
	PageSize int // 每页数量，默认 50
}

// CommentStore 定义评论数据访问接口
type CommentStore interface {
	// CreateComment 创建评论
	CreateComment(ctx context.Context, comment *model.Comment) error
	// GetCommentByID 根据 ID 获取评论
	GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	// GetCommentsByIssueID 获取 Issue 的评论列表（树形结构）
	GetCommentsByIssueID(ctx context.Context, issueID uuid.UUID, opts *ListCommentsOptions) ([]model.Comment, error)
	// GetCommentsByIssueIDWithTotal 获取 Issue 的评论列表（带总数）
	GetCommentsByIssueIDWithTotal(ctx context.Context, issueID uuid.UUID, opts *ListCommentsOptions) ([]model.Comment, int64, error)
	// UpdateComment 更新评论
	UpdateComment(ctx context.Context, id uuid.UUID, body string) error
	// DeleteComment 删除评论
	DeleteComment(ctx context.Context, id uuid.UUID) error
}

// commentStore 实现 CommentStore 接口
type commentStore struct {
	db *gorm.DB
}

// NewCommentStore 创建评论存储实例
func NewCommentStore(db *gorm.DB) CommentStore {
	return &commentStore{db: db}
}

// CreateComment 创建评论
func (s *commentStore) CreateComment(ctx context.Context, comment *model.Comment) error {
	if comment == nil {
		return fmt.Errorf("comment 不能为 nil")
	}

	if err := s.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("创建评论失败: %w", err)
	}

	return nil
}

// GetCommentByID 根据 ID 获取评论
func (s *commentStore) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	var comment model.Comment

	err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Parent").
		First(&comment, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到评论: %w", err)
		}
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}

	return &comment, nil
}

// GetCommentsByIssueID 获取 Issue 的评论列表（树形结构）
func (s *commentStore) GetCommentsByIssueID(ctx context.Context, issueID uuid.UUID, opts *ListCommentsOptions) ([]model.Comment, error) {
	// 默认分页设置
	pageSize := 50
	page := 1
	if opts != nil {
		if opts.PageSize > 0 {
			pageSize = opts.PageSize
		}
		if opts.Page > 0 {
			page = opts.Page
		}
	}

	// 查询所有父评论（parent_id IS NULL）
	var parentComments []model.Comment

	query := s.db.WithContext(ctx).
		Preload("User").
		Where("issue_id = ? AND parent_id IS NULL", issueID).
		Order("created_at ASC")

	// 应用分页
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	if err := query.Find(&parentComments).Error; err != nil {
		return nil, fmt.Errorf("查询评论列表失败: %w", err)
	}

	// 获取每个父评论的回复
	for i := range parentComments {
		var replies []model.Comment
		if err := s.db.WithContext(ctx).
			Preload("User").
			Where("parent_id = ?", parentComments[i].ID).
			Order("created_at ASC").
			Find(&replies).Error; err != nil {
			return nil, fmt.Errorf("查询回复列表失败: %w", err)
		}
		parentComments[i].Replies = replies
	}

	return parentComments, nil
}

// GetCommentsByIssueIDWithTotal 获取 Issue 的评论列表（带总数）
func (s *commentStore) GetCommentsByIssueIDWithTotal(ctx context.Context, issueID uuid.UUID, opts *ListCommentsOptions) ([]model.Comment, int64, error) {
	// 先获取总数（只计算父评论）
	var total int64
	if err := s.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("issue_id = ? AND parent_id IS NULL", issueID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计评论数量失败: %w", err)
	}

	// 获取列表
	comments, err := s.GetCommentsByIssueID(ctx, issueID, opts)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// UpdateComment 更新评论
func (s *commentStore) UpdateComment(ctx context.Context, id uuid.UUID, body string) error {
	now := time.Now()

	result := s.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"body":      body,
			"edited_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("更新评论失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到评论")
	}

	return nil
}

// DeleteComment 删除评论（级联删除回复）
func (s *commentStore) DeleteComment(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&model.Comment{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("删除评论失败: %w", result.Error)
	}

	return nil
}
