// Package handler 提供 HTTP 处理器
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liwei0526vip/mylinear/internal/model"
	"github.com/liwei0526vip/mylinear/internal/service"
)

// TeamMemberHandler 团队成员处理器
type TeamMemberHandler struct {
	teamMemberService service.TeamMemberService
}

// NewTeamMemberHandler 创建团队成员处理器
func NewTeamMemberHandler(teamMemberService service.TeamMemberService) *TeamMemberHandler {
	return &TeamMemberHandler{teamMemberService: teamMemberService}
}

// ListMembers 获取团队成员列表
func (h *TeamMemberHandler) ListMembers(c *gin.Context) {
	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队ID"})
		return
	}

	ctx := contextWithUser(c)

	members, err := h.teamMemberService.ListMembers(ctx, teamID)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]gin.H, len(members))
	for i, member := range members {
		u := gin.H{
			"id":        member.UserID,
			"role":      member.Role,
			"joined_at": member.JoinedAt,
		}
		if member.User != nil {
			u["user"] = gin.H{
				"id":         member.User.ID,
				"name":       member.User.Name,
				"email":      member.User.Email,
				"username":   member.User.Username,
				"avatar_url": member.User.AvatarURL,
			}
		}
		result[i] = u
	}

	c.JSON(http.StatusOK, gin.H{"members": result})
}

// AddMember 添加团队成员
func (h *TeamMemberHandler) AddMember(c *gin.Context) {
	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少团队ID"})
		return
	}

	var req struct {
		UserID string     `json:"user_id" binding:"required"`
		Role   model.Role `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	if req.Role == "" {
		req.Role = model.RoleMember
	}

	ctx := contextWithUser(c)

	err := h.teamMemberService.AddMember(ctx, teamID, req.UserID, req.Role)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "成员已添加"})
}

// RemoveMember 移除团队成员
func (h *TeamMemberHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("teamId")
	userID := c.Param("userId")

	if teamID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	ctx := contextWithUser(c)

	err := h.teamMemberService.RemoveMember(ctx, teamID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成员已移除"})
}

// UpdateMemberRole 更新成员角色
func (h *TeamMemberHandler) UpdateMemberRole(c *gin.Context) {
	teamID := c.Param("teamId")
	userID := c.Param("userId")

	if teamID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	var req struct {
		Role model.Role `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
		return
	}

	ctx := contextWithUser(c)

	err := h.teamMemberService.UpdateRole(ctx, teamID, userID, req.Role)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色已更新"})
}
