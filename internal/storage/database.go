package storage

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(dataDir string) error {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "singbox-web.db")

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// Auto migrate
	err = DB.AutoMigrate(
		&Setting{},
		&Inbound{},
		&Subscription{},
		&Outbound{},
		&Ruleset{},
		&Rule{},
		&OperationLog{},
		&TrafficStat{},
		&RuntimeState{},
	)
	if err != nil {
		return err
	}

	// Initialize default settings
	initDefaultSettings()

	return nil
}

func initDefaultSettings() {
	defaults := map[string]string{
		"web_port":               "60017",
		"username":               "admin",
		"jwt_secret":             generateRandomString(32),
		"singbox_path":           "",
		"download_proxy_enabled": "false",
		"download_proxy_url":     "",
	}

	// Set default password hash (password: 123)
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	defaults["password_hash"] = string(passwordHash)

	for key, value := range defaults {
		var setting Setting
		result := DB.Where("key = ?", key).First(&setting)
		if result.Error == gorm.ErrRecordNotFound {
			DB.Create(&Setting{
				Key:       key,
				Value:     value,
				UpdatedAt: time.Now(),
			})
			log.Printf("Initialized setting: %s", key)
		}
	}
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func GetSetting(key string) (string, error) {
	var setting Setting
	if err := DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func SetSetting(key, value string) error {
	return DB.Save(&Setting{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}).Error
}
