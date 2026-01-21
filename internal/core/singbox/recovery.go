package singbox

import (
	"log"
	"os"
	"strconv"
	"syscall"
	"time"

	"singbox-web/internal/storage"
)

// RecoverFromCrash attempts to recover sing-box state after web service restart
func (m *Manager) RecoverFromCrash() error {
	// Get last status
	lastStatus, err := storage.GetSetting("singbox_status")
	if err != nil || lastStatus != "running" {
		log.Println("Recovery: sing-box was not running before, no recovery needed")
		return nil
	}

	// Get last PID
	pidStr, err := storage.GetSetting("singbox_pid")
	if err != nil {
		log.Println("Recovery: Could not get last PID, will restart sing-box")
		return m.restartSingbox()
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		log.Println("Recovery: Invalid PID, will restart sing-box")
		return m.restartSingbox()
	}

	// Check if process is still alive
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Recovery: Could not find process %d, will restart sing-box", pid)
		return m.restartSingbox()
	}

	// Send signal 0 to check if process exists and we have permission
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		log.Printf("Recovery: Process %d not alive (%v), will restart sing-box", pid, err)
		return m.restartSingbox()
	}

	// Process is still running, reattach
	log.Printf("Recovery: Reattaching to existing sing-box process (PID %d)", pid)

	m.mu.Lock()
	m.status = StatusRunning
	m.mu.Unlock()

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "singbox_recover",
		Detail:    "Reattached to existing sing-box process",
		CreatedAt: time.Now(),
	})

	return nil
}

func (m *Manager) restartSingbox() error {
	log.Println("Recovery: Attempting to restart sing-box...")

	// Check if config exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		log.Println("Recovery: Config file not found, skipping restart")
		storage.SetSetting("singbox_status", "stopped")
		return nil
	}

	// Check if sing-box binary exists
	singboxPath, _ := storage.GetSetting("singbox_path")
	if singboxPath == "" {
		singboxPath = m.dataDir + "/sing-box"
	}
	if _, err := os.Stat(singboxPath); os.IsNotExist(err) {
		log.Println("Recovery: sing-box binary not found, skipping restart")
		storage.SetSetting("singbox_status", "stopped")
		return nil
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "singbox_recover",
		Detail:    "Attempting to restart sing-box after crash",
		CreatedAt: time.Now(),
	})

	return m.Start()
}
