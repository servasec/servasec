package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

var CEF *casbin.Enforcer

const casbinSeedName = "casbin_policies"

func ensureSchemaSeedsTable() {
	DB.Exec(`CREATE TABLE IF NOT EXISTS schema_seeds (
		seed_name  VARCHAR(100) PRIMARY KEY,
		hash       VARCHAR(64)  NOT NULL,
		applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	)`)
}

func csvHash() string {
	data, err := os.ReadFile("config/casbin_policies.csv")
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func needsCasbinSeed() bool {
	currentHash := csvHash()
	if currentHash == "" {
		return false
	}
	var storedHash string
	err := DB.Raw("SELECT hash FROM schema_seeds WHERE seed_name = ?", casbinSeedName).Scan(&storedHash).Error
	if err != nil {
		return true
	}
	return storedHash != currentHash
}

func recordCasbinSeed() {
	currentHash := csvHash()
	if currentHash == "" {
		return
	}
	DB.Exec(`INSERT INTO schema_seeds (seed_name, hash)
		VALUES (?, ?)
		ON CONFLICT (seed_name)
		DO UPDATE SET hash = EXCLUDED.hash, applied_at = NOW()`,
		casbinSeedName, currentHash)
}

func InitCasbin() *casbin.Enforcer {
	adapter, err := gormadapter.NewAdapterByDB(DB)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize casbin adapter: %v", err))
	}

	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		panic(fmt.Sprintf("failed to create casbin enforcer: %v", err))
	}
	enforcer.LoadPolicy()
	CEF = enforcer

	ensureSchemaSeedsTable()
	if needsCasbinSeed() {
		SeedCasbinFromCsv(CEF)
		recordCasbinSeed()
		CEF.LoadPolicy()
	}

	return enforcer
}
