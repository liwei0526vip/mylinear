// Package router 提供路由注册
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mylinear/server/internal/handler"
	"github.com/mylinear/server/internal/middleware"
	"github.com/mylinear/server/internal/service"
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
