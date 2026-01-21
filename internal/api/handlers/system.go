package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"singbox-web/internal/core/singbox"
	"singbox-web/internal/storage"
)

var singboxManager *singbox.Manager
var dataDir string

func InitSystemHandlers(dir string) {
	dataDir = dir
	singboxManager = singbox.GetManager(dataDir)
}

func GetSingboxManager() *singbox.Manager {
	return singboxManager
}

type SystemStatus struct {
	SingboxStatus string `json:"singbox_status"`
	SingboxPid    int    `json:"singbox_pid,omitempty"`
	ConfigExists  bool   `json:"config_exists"`
	SingboxExists bool   `json:"singbox_exists"`
}

func GetSystemStatus(c *gin.Context) {
	status := SystemStatus{
		SingboxStatus: string(singboxManager.GetStatus()),
		SingboxPid:    singboxManager.GetPid(),
	}

	// Check config
	if _, err := os.Stat(singboxManager.GetConfigPath()); err == nil {
		status.ConfigExists = true
	}

	// Check sing-box binary
	singboxPath, _ := storage.GetSetting("singbox_path")
	if singboxPath == "" {
		singboxPath = dataDir + "/sing-box"
	}
	if _, err := os.Stat(singboxPath); err == nil {
		status.SingboxExists = true
	}

	c.JSON(http.StatusOK, status)
}

func StartSingbox(c *gin.Context) {
	if err := singboxManager.Start(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sing-box started"})
}

func StopSingbox(c *gin.Context) {
	if err := singboxManager.Stop(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sing-box stopped"})
}

func RestartSingbox(c *gin.Context) {
	if err := singboxManager.Restart(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sing-box restarted"})
}

func GetSystemVersion(c *gin.Context) {
	version := "0.1.0"

	singboxVersion, err := singbox.GetLatestVersion()
	if err != nil {
		singboxVersion = "unknown"
	}

	c.JSON(http.StatusOK, gin.H{
		"version":                version,
		"singbox_latest_version": singboxVersion,
	})
}

func UpgradeSingbox(c *gin.Context) {
	// Stop sing-box if running
	if singboxManager.GetStatus() == singbox.StatusRunning {
		if err := singboxManager.Stop(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop sing-box: " + err.Error()})
			return
		}
	}

	version, err := singbox.DownloadLatest(dataDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "singbox_upgrade",
		Detail:    "Upgraded to " + version,
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "sing-box upgraded",
		"version": version,
	})
}

func GetGeneratedConfig(c *gin.Context) {
	configPath := singboxManager.GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}
