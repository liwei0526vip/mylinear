package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/liwei0526vip/mylinear/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testDB is used by Workflow tests
var testDB *gorm.DB

// testSvcDB is used by service tests
var testSvcDB *gorm.DB

// TestMain initializes the DB connection for all service tests
func TestMain(m *testing.M) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Use same test DB as store package (mylinear credentials)
		databaseURL = "postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable"
	}

	var err error
	testDB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("无法连接数据库: %v\n", err)
		os.Exit(1)
	}

	// Support existing tests that rely on testTeamServiceDB
	testTeamServiceDB = testDB

	// Support Issue service tests
	testSvcDB = testDB

	// 统一清理和迁移
	testDB.Exec("DROP TABLE IF EXISTS issue_subscriptions CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS issues CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS labels CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS workflow_states CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS team_members CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS teams CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS users CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS workspaces CASCADE")

	err = testDB.AutoMigrate(
		&model.Workspace{},
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.WorkflowState{},
		&model.Label{},
		&model.Issue{},
		&model.IssueSubscription{},
	)
	if err != nil {
		fmt.Printf("自动迁移失败: %v\n", err)
		os.Exit(1)
	}

	// Set Store's test DB if needed? No, store package is imported strictly for types?
	// But team_test.go uses store.NewTeamStore(tx). Correct.

	// Support existing tests that rely on testTeamServiceDB variable
	testTeamServiceDB = testDB

	os.Exit(m.Run())
}
