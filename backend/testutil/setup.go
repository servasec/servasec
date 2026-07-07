package testutil

import (
	"os"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func init() {
	gin.SetMode(gin.TestMode)
}

func SetupTestDB() *gorm.DB {
	if testDB != nil {
		return testDB
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database: " + err.Error())
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.BlacklistedToken{},
		&models.Group{},
		&models.Application{},
		&models.ApplicationVersion{},
		&models.ScannerType{},
		&models.Scan{},
		&models.Finding{},
		&models.Team{},
		&models.TeamMember{},
		&models.Comment{},
		&models.Webhook{},
		&models.AuditLog{},
		&models.Policy{},
		&models.PolicyLog{},
		&models.UserApiKey{},
	)
	if err != nil {
		panic("failed to migrate test database: " + err.Error())
	}

	config.DB = db
	testDB = db
	return db
}

func SetupTestCasbin() *casbin.Enforcer {
	wd, _ := os.Getwd()
	if wd != "" {
		os.Chdir(wd)
	}

	adapter, err := gormadapter.NewAdapterByDB(config.DB)
	if err != nil {
		panic("failed to create casbin adapter: " + err.Error())
	}

	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		panic("failed to create casbin enforcer: " + err.Error())
	}
	enforcer.LoadPolicy()

	config.CEF = enforcer
	return enforcer
}

func SeedTestUser(db *gorm.DB, username, email, password, role string) *models.User {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashed),
		Role:     role,
	}
	if err := db.Create(user).Error; err != nil {
		panic("failed to seed test user: " + err.Error())
	}
	return user
}

func SeedTestScannerType(db *gorm.DB, name, parser string) *models.ScannerType {
	st := &models.ScannerType{
		Name:   name,
		Parser: parser,
	}
	if err := db.Create(st).Error; err != nil {
		panic("failed to seed scanner type: " + err.Error())
	}
	return st
}

func Now() time.Time {
	return time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
}
