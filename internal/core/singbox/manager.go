package singbox

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"singbox-web/internal/storage"
)

type Status string

const (
	StatusStopped Status = "stopped"
	StatusRunning Status = "running"
	StatusError   Status = "error"
)

type Manager struct {
	mu         sync.RWMutex
	cmd        *exec.Cmd
	status     Status
	dataDir    string
	configPath string
	logChan    chan string
	stopChan   chan struct{}
}

var instance *Manager
var once sync.Once

func GetManager(dataDir string) *Manager {
	once.Do(func() {
		instance = &Manager{
			status:     StatusStopped,
			dataDir:    dataDir,
			configPath: filepath.Join(dataDir, "config.json"),
			logChan:    make(chan string, 1000),
		}
	})
	return instance
}

func (m *Manager) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *Manager) GetLogChannel() <-chan string {
	return m.logChan
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == StatusRunning {
		return fmt.Errorf("sing-box is already running")
	}

	// Get sing-box path
	singboxPath, err := storage.GetSetting("singbox_path")
	if err != nil || singboxPath == "" {
		singboxPath = filepath.Join(m.dataDir, "sing-box")
	}

	// Check if sing-box binary exists
	if _, err := os.Stat(singboxPath); os.IsNotExist(err) {
		return fmt.Errorf("sing-box binary not found at %s, please download first", singboxPath)
	}

	// Check if config exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", m.configPath)
	}

	// Start sing-box
	m.cmd = exec.Command(singboxPath, "run", "-c", m.configPath)
	m.cmd.Dir = m.dataDir

	// Capture stdout and stderr
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := m.cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	m.status = StatusRunning
	m.stopChan = make(chan struct{})

	// Save runtime state
	storage.SetSetting("singbox_status", "running")
	storage.SetSetting("singbox_pid", fmt.Sprintf("%d", m.cmd.Process.Pid))
	storage.SetSetting("last_start_time", time.Now().Format(time.RFC3339))

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "singbox_start",
		Detail:    fmt.Sprintf("sing-box started with PID %d", m.cmd.Process.Pid),
		CreatedAt: time.Now(),
	})

	// Read logs in goroutines
	go m.readLogs(stdout)
	go m.readLogs(stderr)

	// Wait for process in background
	go func() {
		err := m.cmd.Wait()
		m.mu.Lock()
		if m.status == StatusRunning {
			if err != nil {
				m.status = StatusError
				select {
				case m.logChan <- fmt.Sprintf("sing-box exited with error: %v", err):
				default:
				}
			} else {
				m.status = StatusStopped
			}
		}
		storage.SetSetting("singbox_status", string(m.status))
		close(m.stopChan)
		m.mu.Unlock()
	}()

	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != StatusRunning {
		return fmt.Errorf("sing-box is not running")
	}

	// Send SIGTERM
	if m.cmd != nil && m.cmd.Process != nil {
		if err := m.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			// Force kill if SIGTERM fails
			m.cmd.Process.Kill()
		}
	}

	m.status = StatusStopped
	storage.SetSetting("singbox_status", "stopped")

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action:    "singbox_stop",
		Detail:    "sing-box stopped",
		CreatedAt: time.Now(),
	})

	return nil
}

func (m *Manager) Restart() error {
	if m.GetStatus() == StatusRunning {
		if err := m.Stop(); err != nil {
			return err
		}
		// Wait a bit for process to fully stop
		time.Sleep(time.Second)
	}
	return m.Start()
}

func (m *Manager) readLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case m.logChan <- line:
		default:
			// Drop log if channel full
		}
	}
}

func (m *Manager) GetPid() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Pid
	}
	return 0
}

func (m *Manager) GetConfigPath() string {
	return m.configPath
}

func (m *Manager) GetDataDir() string {
	return m.dataDir
}
