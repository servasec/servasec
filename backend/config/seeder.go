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
	if os.Getenv("SSC_DEBUG_ENABLED") != "true" {
		log.Fatal("SSC_ADMIN_PASSWORD must be set in production")
	}
	randomPwd, err := generateRandomHex(16)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to generate random admin password: %v. refusing to fall back to hardcoded password.", err)
	}

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
	csvPolicies, err := csvEnforcer.GetPolicy()
	if err != nil {
		debug.Log("Failed to get CSV policies: %v", err)
		return
	}

	added := 0
	for _, p := range csvPolicies {
		if len(p) >= 3 {
			ok, _ := enforcer.AddPolicy(p[0], p[1], p[2])
			if ok {
				added++
			}
		}
	}

	if added > 0 {
		if err := enforcer.SavePolicy(); err != nil {
			debug.Log("Failed to save casbin policies: %v", err)
			return
		}
	}
	debug.Log("Seeding: Casbin from CSV finished (%d new policies)", added)
}

func seedScannerTypes() {
	scannerTypes := []models.ScannerType{
		{Name: "semgrep", Description: "Semgrep SAST (SARIF/JSON)", Parser: "semgrep", Enabled: true},
		{Name: "trivy", Description: "Trivy vulnerability scanner (SARIF/JSON)", Parser: "trivy", Enabled: true},
		{Name: "gitleaks", Description: "Gitleaks secret detection (JSON)", Parser: "gitleaks", Enabled: true},
		{Name: "grype", Description: "Grype vulnerability scanner (JSON)", Parser: "grype", Enabled: true},
		{Name: "snyk", Description: "Snyk (SARIF/JSON)", Parser: "snyk", Enabled: true},
		{Name: "checkov", Description: "Checkov IaC scan (SARIF)", Parser: "checkov", Enabled: true},
		{Name: "trufflehog", Description: "TruffleHog secret detection (JSON)", Parser: "trufflehog", Enabled: true},
		{Name: "nuclei", Description: "Nuclei DAST/template scanner (JSON/JSONL)", Parser: "nuclei", Enabled: true},
		{Name: "sarif", Description: "Generic SARIF v2.1.0 parser (fallback for unknown tools)", Parser: "sarif", Enabled: true},
		{Name: "gosec", Description: "Gosec Go SAST (JSON)", Parser: "gosec", Enabled: true},
		{Name: "bandit", Description: "Bandit Python SAST (JSON)", Parser: "bandit", Enabled: true},
		{Name: "osv-scanner", Description: "OSV-Scanner SCA (JSON)", Parser: "osv-scanner", Enabled: true},
		{Name: "npm-audit", Description: "NPM Audit SCA (JSON)", Parser: "npm-audit", Enabled: true},
		{Name: "tfsec", Description: "Tfsec Terraform IaC (JSON)", Parser: "tfsec", Enabled: true},
		{Name: "kubescape", Description: "Kubescape Kubernetes CSPM (JSON)", Parser: "kubescape", Enabled: true},
		{Name: "kube-bench", Description: "Kube-bench CIS Kubernetes (JSON)", Parser: "kube-bench", Enabled: true},
	}

	for _, st := range scannerTypes {
		var existing models.ScannerType
		err := DB.Where("name = ?", st.Name).First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			debug.Log("Failed to check scanner type %s: %v\n", st.Name, err)
			continue
		}
		if err == nil {
			if !existing.Enabled {
				existing.Enabled = true
				DB.Save(&existing)
			}
			continue
		}
		if err := DB.Create(&st).Error; err != nil {
			debug.Log("Failed to seed scanner type %s: %v\n", st.Name, err)
		}
	}
	debug.Println("Seeding: scanner types finished")
}

func SeedDatabase() {
	debug.Println("Seeding: Database..")
	seedDefaultUsers()
	seedScannerTypes()
}
