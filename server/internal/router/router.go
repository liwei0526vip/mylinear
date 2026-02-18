// Package router 提供路由注册
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/liwei0526vip/mylinear/internal/handler"
	"github.com/liwei0526vip/mylinear/internal/middleware"
	"github.com/liwei0526vip/mylinear/internal/service"
	"github.com/liwei0526vip/mylinear/internal/store"
	"gorm.io/gorm"
)

// RegisterWorkspaceRoutes 注册 Workspace 路由
func RegisterWorkspaceRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtService service.JWTService, workspaceService service.WorkspaceService) {
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)

	workspaceGroup := rg.Group("/workspaces")
	workspaceGroup.Use(func(c *gin.Context) {
		c.Set("db", db)
	})
	workspaceGroup.Use(middleware.Auth(jwtService))
	{
		workspaceGroup.GET("/:id", workspaceHandler.GetWorkspace)
		workspaceGroup.PUT("/:id", workspaceHandler.UpdateWorkspace)
	}
}

// RegisterTeamRoutes 注册 Team 路由
func RegisterTeamRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtService service.JWTService, teamService service.TeamService) {
	teamHandler := handler.NewTeamHandler(teamService)

	teamsGroup := rg.Group("/teams")
	teamsGroup.Use(func(c *gin.Context) {
		c.Set("db", db)
	})
	teamsGroup.Use(middleware.Auth(jwtService))
	{
		teamsGroup.GET("", teamHandler.ListTeams)
		teamsGroup.POST("", teamHandler.CreateTeam)
		teamsGroup.GET("/:teamId", teamHandler.GetTeam)
		teamsGroup.PUT("/:teamId", teamHandler.UpdateTeam)
		teamsGroup.DELETE("/:teamId", teamHandler.DeleteTeam)
	}
}

// RegisterTeamMemberRoutes 注册 TeamMember 路由
func RegisterTeamMemberRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtService service.JWTService, teamMemberService service.TeamMemberService) {
	teamMemberHandler := handler.NewTeamMemberHandler(teamMemberService)

	teamsGroup := rg.Group("/teams")
	teamsGroup.Use(func(c *gin.Context) {
		c.Set("db", db)
	})
	teamsGroup.Use(middleware.Auth(jwtService))
	{
		teamsGroup.GET("/:teamId/members", teamMemberHandler.ListMembers)
		teamsGroup.POST("/:teamId/members", teamMemberHandler.AddMember)
		teamsGroup.DELETE("/:teamId/members/:userId", teamMemberHandler.RemoveMember)
		teamsGroup.PUT("/:teamId/members/:userId", teamMemberHandler.UpdateMemberRole)
	}
}

// RegisterWorkflowRoutes 注册 Workflow 路由
func RegisterWorkflowRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtService service.JWTService, workflowService service.WorkflowService) {
	workflowHandler := handler.NewWorkflowHandler(workflowService)

	workflowGroup := rg.Group("")
	workflowGroup.Use(func(c *gin.Context) {
		c.Set("db", db)
	})
	workflowGroup.Use(middleware.Auth(jwtService))
	{
		workflowGroup.GET("/teams/:teamId/workflow-states", workflowHandler.ListStates)
		workflowGroup.POST("/teams/:teamId/workflow-states", workflowHandler.CreateState)
		workflowGroup.PUT("/workflow-states/:id", workflowHandler.UpdateState)
		workflowGroup.DELETE("/workflow-states/:id", workflowHandler.DeleteState)
	}
}

// RegisterLabelRoutes 注册 Label 路由
func RegisterLabelRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtService service.JWTService, labelService service.LabelService, teamStore store.TeamStore) {
	labelHandler := handler.NewLabelHandler(labelService, teamStore)

	labelGroup := rg.Group("")
	labelGroup.Use(func(c *gin.Context) {
		c.Set("db", db)
	})
	labelGroup.Use(middleware.Auth(jwtService))
	{
		labelGroup.GET("/teams/:teamId/labels", labelHandler.ListLabels)
		labelGroup.POST("/teams/:teamId/labels", labelHandler.CreateLabel)
	}
}

// RegisterIssueRoutes 注册 Issue 路由
func RegisterIssueRoutes(rg *gin.RouterGroup, db *gorm.DB, jwtService service.JWTService, issueService service.IssueService) {
	issueHandler := handler.NewIssueHandler(issueService)

	issueGroup := rg.Group("")
	issueGroup.Use(func(c *gin.Context) {
		c.Set("db", db)
	})
	issueGroup.Use(middleware.Auth(jwtService))
	{
		// 团队内 Issue 操作
		issueGroup.GET("/teams/:teamId/issues", issueHandler.ListIssues)
		issueGroup.POST("/teams/:teamId/issues", issueHandler.CreateIssue)

		// Issue CRUD
		issueGroup.GET("/issues/:id", issueHandler.GetIssue)
		issueGroup.PUT("/issues/:id", issueHandler.UpdateIssue)
		issueGroup.DELETE("/issues/:id", issueHandler.DeleteIssue)

		// Issue 位置更新
		issueGroup.PUT("/issues/:id/position", issueHandler.UpdatePosition)

		// Issue 订阅
		issueGroup.POST("/issues/:id/subscribe", issueHandler.Subscribe)
		issueGroup.DELETE("/issues/:id/subscribe", issueHandler.Unsubscribe)
		issueGroup.GET("/issues/:id/subscribers", issueHandler.ListSubscribers)

		// Issue 恢复
		issueGroup.POST("/issues/:id/restore", issueHandler.RestoreIssue)
	}
}
