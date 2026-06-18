package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func generateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getAdminPassword() string {
	if pwd := os.Getenv("SSC_ADMIN_PASSWORD"); pwd != "" {
		return pwd
	}
	randomPwd, err := generateRandomHex(16)
	if err != nil {
		log.Printf("Failed to generate random admin password: %v", err)
		return "Admin1234!"
	}

	// use log package instead of custom debug so even if debug is disabled, admin see its generated password
	log.Printf("========================================")
	log.Printf("  ADMIN PASSWORD: %s", randomPwd)
	log.Printf("  Set SSC_ADMIN_PASSWORD env var to disable random generation")
	log.Printf("========================================")
	return randomPwd
}

func seedDefaultUsers() {
	adminPassword := getAdminPassword()

	users := []models.User{
		{Username: "admin", Email: "admin@servasec.local", Password: adminPassword, Role: "admin"},
	}

	for _, user := range users {
		var existing models.User
		err := DB.Where("username = ? OR email = ?", user.Username, user.Email).First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			debug.Log("Failed to check user %s: %v\n", user.Username, err)
			continue
		}
		if err == nil {
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			debug.Log("Failed to hash password for user %s: %v\n", user.Username, err)
			continue
		}
		user.Password = string(hashedPassword)
		if err := DB.Create(&user).Error; err != nil {
			debug.Log("Failed to seed user %s: %v\n", user.Username, err)
		}
	}
	debug.Println("Seeding: users finished")
}

func SeedCasbinFromCsv(enforcer *casbin.Enforcer) {
	debug.Println("Seeding: Casbin rules from CSV..")
	csvEnforcer, err := casbin.NewEnforcer("config/casbin_model.conf", "config/casbin_policies.csv")
	if err != nil {
		debug.Println(err.Error())
		return
	}
	csvEnforcer.LoadPolicy()
	csvEnforcer.SetAdapter(enforcer.GetAdapter())
	err = csvEnforcer.SavePolicy()
	if err != nil {
		debug.Log("Failed to save casbin policies: %v", err)
		return
	}
	debug.Println("Seeding: Casbin from CSV finished")
}

func SeedDatabase() {
	debug.Println("Seeding: Database..")
	seedDefaultUsers()
}
