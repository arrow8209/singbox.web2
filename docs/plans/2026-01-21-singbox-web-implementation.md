# singbox-web 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个 sing-box Web 管理客户端，支持中转服务器场景的完整功能

**Architecture:** Go 后端 + Vue 3 前端单体架构，SQLite 数据存储，JWT 认证，WebSocket 实时日志

**Tech Stack:** Go 1.21+, Gin, GORM, SQLite, Vue 3, Vite, Element Plus, Pinia

---

## 里程碑概览

| 里程碑 | 内容 | 预期产出 |
|-------|------|---------|
| M1 | 项目初始化 | Go + Vue 项目骨架，可运行 |
| M2 | 认证系统 | 登录/登出，JWT 认证 |
| M3 | sing-box 管理 | 下载、启停、进程管理 |
| M4 | 入站管理 | 入站 CRUD + 前端页面 |
| M5 | 出站/订阅管理 | 订阅解析、节点管理、测速 |
| M6 | 规则管理 | 规则集、路由规则、规则中心 |
| M7 | 配置生成 | 完整配置生成、应用配置 |
| M8 | 定时任务 | 订阅/规则集自动更新 |
| M9 | 日志系统 | 操作日志、实时日志 WebSocket |
| M10 | 集成构建 | 前端嵌入、单二进制构建 |

---

## M1: 项目初始化

### Task 1.1: 初始化 Go 项目

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `cmd/singbox-web/main.go`

**Step 1: 初始化 Go module**

```bash
cd /home/zhuyb/Documents/1.code/singbox.arrow.web2
go mod init singbox-web
```

**Step 2: 创建入口文件**

Create `cmd/singbox-web/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := 60017

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("singbox-web starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
```

**Step 3: 运行验证**

```bash
go run cmd/singbox-web/main.go &
sleep 2
curl http://localhost:60017/api/health
# Expected: {"status":"ok"}
pkill -f "singbox-web"
```

**Step 4: 提交**

```bash
git add go.mod cmd/
git commit -m "feat: initialize Go project with health endpoint"
```

---

### Task 1.2: 创建目录结构

**Files:**
- Create: `internal/api/router.go`
- Create: `internal/api/middleware/.gitkeep`
- Create: `internal/api/handlers/.gitkeep`
- Create: `internal/core/singbox/.gitkeep`
- Create: `internal/core/parser/.gitkeep`
- Create: `internal/core/generator/.gitkeep`
- Create: `internal/core/scheduler/.gitkeep`
- Create: `internal/storage/.gitkeep`
- Create: `internal/websocket/.gitkeep`
- Create: `configs/.gitkeep`
- Create: `scripts/.gitkeep`
- Create: `.gitignore`

**Step 1: 创建目录和占位文件**

```bash
mkdir -p internal/api/middleware internal/api/handlers
mkdir -p internal/core/singbox internal/core/parser internal/core/generator internal/core/scheduler
mkdir -p internal/storage internal/websocket
mkdir -p configs scripts data

touch internal/api/middleware/.gitkeep
touch internal/api/handlers/.gitkeep
touch internal/core/singbox/.gitkeep
touch internal/core/parser/.gitkeep
touch internal/core/generator/.gitkeep
touch internal/core/scheduler/.gitkeep
touch internal/storage/.gitkeep
touch internal/websocket/.gitkeep
touch configs/.gitkeep
touch scripts/.gitkeep
```

**Step 2: 创建 .gitignore**

Create `.gitignore`:

```
# Build output
/singbox-web
*.exe

# Data directory
/data/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Frontend
/frontend/node_modules/
/frontend/dist/

# Test
coverage.out
*.test

# Temp
*.tmp
*.log
```

**Step 3: 提交**

```bash
git add .
git commit -m "feat: create project directory structure"
```

---

### Task 1.3: 安装后端依赖

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: 添加核心依赖**

```bash
go get github.com/gin-gonic/gin@v1.9.1
go get gorm.io/gorm@v1.25.5
go get gorm.io/driver/sqlite@v1.5.4
go get github.com/golang-jwt/jwt/v5@v5.2.0
go get github.com/gorilla/websocket@v1.5.1
go get golang.org/x/crypto@latest
```

**Step 2: 整理依赖**

```bash
go mod tidy
```

**Step 3: 提交**

```bash
git add go.mod go.sum
git commit -m "feat: add backend dependencies"
```

---

### Task 1.4: 初始化 Vue 3 前端项目

**Files:**
- Create: `frontend/` (Vue 3 project)

**Step 1: 创建 Vue 项目**

```bash
cd /home/zhuyb/Documents/1.code/singbox.arrow.web2
npm create vite@latest frontend -- --template vue-ts
```

**Step 2: 安装前端依赖**

```bash
cd frontend
npm install
npm install vue-router@4 pinia axios element-plus @element-plus/icons-vue
npm install -D sass unplugin-auto-import unplugin-vue-components
cd ..
```

**Step 3: 验证前端运行**

```bash
cd frontend
npm run dev &
sleep 3
curl -s http://localhost:5173 | head -20
pkill -f "vite"
cd ..
```

**Step 4: 提交**

```bash
git add frontend/
git commit -m "feat: initialize Vue 3 frontend project"
```

---

### Task 1.5: 配置 Vite 和 Element Plus

**Files:**
- Modify: `frontend/vite.config.ts`
- Modify: `frontend/src/main.ts`

**Step 1: 更新 vite.config.ts**

Replace `frontend/vite.config.ts`:

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import path from 'path'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:60017',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:60017',
        ws: true,
      },
    },
  },
})
```

**Step 2: 更新 main.ts**

Replace `frontend/src/main.ts`:

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'

const app = createApp(App)

// Register all Element Plus icons
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
```

**Step 3: 创建路由配置**

Create `frontend/src/router/index.ts`:

```typescript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
  ],
})

export default router
```

**Step 4: 创建首页占位**

Create `frontend/src/views/HomeView.vue`:

```vue
<template>
  <div class="home">
    <h1>singbox-web</h1>
    <el-button type="primary">Element Plus Ready</el-button>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 20px;
}
</style>
```

**Step 5: 验证配置**

```bash
cd frontend
npm run dev &
sleep 5
curl -s http://localhost:5173 | grep -o "singbox-web" || echo "Check manually at http://localhost:5173"
pkill -f "vite"
cd ..
```

**Step 6: 提交**

```bash
git add frontend/
git commit -m "feat: configure Vite, Element Plus and Vue Router"
```

---

## M2: 认证系统

### Task 2.1: 创建数据库模型和初始化

**Files:**
- Create: `internal/storage/database.go`
- Create: `internal/storage/models.go`

**Step 1: 创建数据库模型**

Create `internal/storage/models.go`:

```go
package storage

import (
	"time"
)

type Setting struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedAt time.Time
}

type Inbound struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Type      string `gorm:"not null"`
	Config    string `gorm:"not null"` // JSON
	Enabled   bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Subscription struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	URL            string `gorm:"not null"`
	Type           string // auto/singbox/clash/v2ray/base64
	UpdateInterval int    // hours
	LastUpdate     *time.Time
	Enabled        bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Outbound struct {
	ID             uint   `gorm:"primaryKey"`
	SubscriptionID *uint  // NULL for manual
	Name           string `gorm:"not null"`
	Type           string `gorm:"not null"`
	Server         string
	Port           int
	Config         string `gorm:"not null"` // JSON
	Latency        *int   // ms
	Enabled        bool   `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Subscription *Subscription `gorm:"foreignKey:SubscriptionID"`
}

type Ruleset struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	Type           string `gorm:"not null"` // remote/local
	Format         string `gorm:"not null"` // source/binary
	URL            string
	Path           string // local file path
	UpdateInterval int    // hours
	LastUpdate     *time.Time
	Enabled        bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Rule struct {
	ID          uint   `gorm:"primaryKey"`
	Priority    int    `gorm:"not null"`
	Type        string `gorm:"not null"` // domain/ip/geoip/geosite/ruleset...
	Value       string `gorm:"not null"`
	OutboundTag string `gorm:"not null"`
	Enabled     bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OperationLog struct {
	ID        uint   `gorm:"primaryKey"`
	Action    string `gorm:"not null"`
	Detail    string
	CreatedAt time.Time
}

type TrafficStat struct {
	ID         uint   `gorm:"primaryKey"`
	TargetType string `gorm:"not null"` // outbound/inbound/rule
	TargetName string `gorm:"not null"`
	Upload     int64  `gorm:"default:0"`
	Download   int64  `gorm:"default:0"`
	Date       string `gorm:"not null"` // YYYY-MM-DD
}

type RuntimeState struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedAt time.Time
}
```

**Step 2: 创建数据库初始化**

Create `internal/storage/database.go`:

```go
package storage

import (
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

	// Set default password hash
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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
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
```

**Step 3: 验证编译**

```bash
go build ./...
```

**Step 4: 提交**

```bash
git add internal/storage/
git commit -m "feat: add database models and initialization"
```

---

### Task 2.2: 实现 JWT 认证中间件

**Files:**
- Create: `internal/api/middleware/auth.go`

**Step 1: 创建认证中间件**

Create `internal/api/middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"singbox-web/internal/storage"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		jwtSecret, err := storage.GetSetting("jwt_secret")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get jwt secret"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

func GenerateToken(username string) (string, error) {
	jwtSecret, err := storage.GetSetting("jwt_secret")
	if err != nil {
		return "", err
	}

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
```

**Step 2: 验证编译**

```bash
go build ./...
```

**Step 3: 提交**

```bash
git add internal/api/middleware/
git commit -m "feat: add JWT authentication middleware"
```

---

### Task 2.3: 实现认证 API

**Files:**
- Create: `internal/api/handlers/auth.go`

**Step 1: 创建认证处理器**

Create `internal/api/handlers/auth.go`:

```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"singbox-web/internal/api/middleware"
	"singbox-web/internal/storage"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=3"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get stored credentials
	storedUsername, err := storage.GetSetting("username")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}

	storedPasswordHash, err := storage.GetSetting("password_hash")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get password"})
		return
	}

	// Verify credentials
	if req.Username != storedUsername {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "login",
		Detail: "User logged in: " + req.Username,
	})

	c.JSON(http.StatusOK, LoginResponse{
		Token:    token,
		Username: req.Username,
	})
}

func Logout(c *gin.Context) {
	username := c.GetString("username")

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "logout",
		Detail: "User logged out: " + username,
	})

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify old password
	storedPasswordHash, err := storage.GetSetting("password_hash")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Update password
	if err := storage.SetSetting("password_hash", string(newHash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "change_password",
		Detail: "Password changed",
	})

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}
```

**Step 2: 验证编译**

```bash
go build ./...
```

**Step 3: 提交**

```bash
git add internal/api/handlers/auth.go
git commit -m "feat: add authentication API handlers"
```

---

### Task 2.4: 创建路由配置

**Files:**
- Create: `internal/api/router.go`
- Modify: `cmd/singbox-web/main.go`

**Step 1: 创建路由配置**

Create `internal/api/router.go`:

```go
package api

import (
	"github.com/gin-gonic/gin"
	"singbox-web/internal/api/handlers"
	"singbox-web/internal/api/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth
			protected.POST("/auth/logout", handlers.Logout)
			protected.PUT("/auth/password", handlers.ChangePassword)

			// TODO: Add more routes
		}
	}

	return r
}
```

**Step 2: 更新 main.go**

Replace `cmd/singbox-web/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"singbox-web/internal/api"
	"singbox-web/internal/storage"
)

func main() {
	// Get data directory
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := filepath.Join(filepath.Dir(execPath), "data")

	// Use current directory for development
	if os.Getenv("DEV") == "1" {
		dataDir = "data"
	}

	// Initialize database
	if err := storage.InitDatabase(dataDir); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Get port from settings
	port, err := storage.GetSetting("web_port")
	if err != nil {
		port = "60017"
	}

	// Setup router
	r := api.SetupRouter()

	// Start server
	log.Printf("singbox-web starting on port %s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}
```

**Step 3: 运行测试**

```bash
DEV=1 go run cmd/singbox-web/main.go &
sleep 3

# Test health
curl http://localhost:60017/api/health
# Expected: {"status":"ok"}

# Test login
curl -X POST http://localhost:60017/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123"}'
# Expected: {"token":"...", "username":"admin"}

pkill -f "singbox-web"
```

**Step 4: 提交**

```bash
git add internal/api/router.go cmd/singbox-web/main.go
git commit -m "feat: setup Gin router with auth endpoints"
```

---

### Task 2.5: 创建前端登录页面

**Files:**
- Create: `frontend/src/api/auth.ts`
- Create: `frontend/src/stores/user.ts`
- Create: `frontend/src/views/LoginView.vue`
- Modify: `frontend/src/router/index.ts`

**Step 1: 创建 API 客户端**

Create `frontend/src/api/index.ts`:

```typescript
import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Request interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    ElMessage.error(error.response?.data?.error || 'Request failed')
    return Promise.reject(error)
  }
)

export default api
```

**Step 2: 创建认证 API**

Create `frontend/src/api/auth.ts`:

```typescript
import api from './index'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  username: string
}

export const authApi = {
  login(data: LoginRequest) {
    return api.post<LoginResponse>('/auth/login', data)
  },
  logout() {
    return api.post('/auth/logout')
  },
  changePassword(oldPassword: string, newPassword: string) {
    return api.put('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },
}
```

**Step 3: 创建用户 Store**

Create `frontend/src/stores/user.ts`:

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type LoginRequest } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')

  const isLoggedIn = computed(() => !!token.value)

  async function login(data: LoginRequest) {
    const response = await authApi.login(data)
    token.value = response.data.token
    username.value = response.data.username
    localStorage.setItem('token', response.data.token)
    localStorage.setItem('username', response.data.username)
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      token.value = ''
      username.value = ''
      localStorage.removeItem('token')
      localStorage.removeItem('username')
    }
  }

  return {
    token,
    username,
    isLoggedIn,
    login,
    logout,
  }
})
```

**Step 4: 创建登录页面**

Create `frontend/src/views/LoginView.vue`:

```vue
<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <h2>singbox-web</h2>
      </template>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleLogin"
      >
        <el-form-item label="Username" prop="username">
          <el-input v-model="form.username" placeholder="Enter username" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="Enter password"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            native-type="submit"
            :loading="loading"
            class="login-button"
          >
            Login
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: 'Please enter username', trigger: 'blur' }],
  password: [{ required: true, message: 'Please enter password', trigger: 'blur' }],
}

async function handleLogin() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      await userStore.login(form)
      ElMessage.success('Login successful')
      router.push('/')
    } catch (error) {
      // Error handled by interceptor
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #f5f7fa;
}

.login-card {
  width: 400px;
}

.login-card h2 {
  margin: 0;
  text-align: center;
}

.login-button {
  width: 100%;
}
</style>
```

**Step 5: 更新路由**

Replace `frontend/src/router/index.ts`:

```typescript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')

  if (to.meta.requiresAuth !== false && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/')
  } else {
    next()
  }
})

export default router
```

**Step 6: 提交**

```bash
git add frontend/src/
git commit -m "feat: add login page and auth store"
```

---

## M3: sing-box 管理

### Task 3.1: 实现 sing-box 下载器

**Files:**
- Create: `internal/core/singbox/downloader.go`

**Step 1: 创建下载器**

Create `internal/core/singbox/downloader.go`:

```go
package singbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"singbox-web/internal/storage"
)

const (
	GitHubAPIURL = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getHTTPClient() *http.Client {
	client := &http.Client{}

	proxyEnabled, _ := storage.GetSetting("download_proxy_enabled")
	proxyURL, _ := storage.GetSetting("download_proxy_url")

	if proxyEnabled == "true" && proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	return client
}

func GetLatestVersion() (string, error) {
	client := getHTTPClient()

	resp, err := client.Get(GitHubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
}

func DownloadLatest(dataDir string) (string, error) {
	client := getHTTPClient()

	// Get release info
	resp, err := client.Get(GitHubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	// Find matching asset
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	}

	expectedName := fmt.Sprintf("sing-box-%s-linux-%s.tar.gz",
		strings.TrimPrefix(release.TagName, "v"), arch)

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no matching asset found for linux-%s", arch)
	}

	// Download
	resp, err = client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Save to temp file
	tmpFile := filepath.Join(dataDir, "sing-box.tar.gz")
	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	// Extract
	binPath := filepath.Join(dataDir, "sing-box")
	if err := extractTarGz(tmpFile, dataDir, "sing-box"); err != nil {
		return "", err
	}

	// Make executable
	if err := os.Chmod(binPath, 0755); err != nil {
		return "", err
	}

	// Cleanup
	os.Remove(tmpFile)

	// Save path to settings
	storage.SetSetting("singbox_path", binPath)

	return release.TagName, nil
}

func extractTarGz(tarGzPath, destDir, targetFile string) error {
	// Use tar command for simplicity
	cmd := fmt.Sprintf("tar -xzf %s -C %s --strip-components=1 --wildcards '*/%s'",
		tarGzPath, destDir, targetFile)

	return runCommand("sh", "-c", cmd)
}

func runCommand(name string, args ...string) error {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err == nil {
			cmd.Path = lp
		}
	}
	return cmd.Run()
}
```

**Step 2: 添加缺失的 import**

Add to imports:

```go
import (
	"os/exec"
)
```

**Step 3: 验证编译**

```bash
go build ./...
```

**Step 4: 提交**

```bash
git add internal/core/singbox/downloader.go
git commit -m "feat: add sing-box downloader with proxy support"
```

---

### Task 3.2: 实现 sing-box 进程管理

**Files:**
- Create: `internal/core/singbox/manager.go`

**Step 1: 创建进程管理器**

Create `internal/core/singbox/manager.go`:

```go
package singbox

import (
	"bufio"
	"fmt"
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
		return fmt.Errorf("sing-box binary not found, please download first")
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
		Action: "singbox_start",
		Detail: fmt.Sprintf("sing-box started with PID %d", m.cmd.Process.Pid),
	})

	// Read logs in goroutines
	go m.readLogs(stdout)
	go m.readLogs(stderr)

	// Wait for process
	go func() {
		err := m.cmd.Wait()
		m.mu.Lock()
		if m.status == StatusRunning {
			if err != nil {
				m.status = StatusError
				m.logChan <- fmt.Sprintf("sing-box exited with error: %v", err)
			} else {
				m.status = StatusStopped
			}
		}
		storage.SetSetting("singbox_status", string(m.status))
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
			// Force kill
			m.cmd.Process.Kill()
		}
	}

	// Wait for stop with timeout
	select {
	case <-m.stopChan:
	case <-time.After(5 * time.Second):
		if m.cmd != nil && m.cmd.Process != nil {
			m.cmd.Process.Kill()
		}
	}

	m.status = StatusStopped
	storage.SetSetting("singbox_status", "stopped")

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "singbox_stop",
		Detail: "sing-box stopped",
	})

	return nil
}

func (m *Manager) Restart() error {
	if m.GetStatus() == StatusRunning {
		if err := m.Stop(); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return m.Start()
}

func (m *Manager) readLogs(reader *os.File) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case m.logChan <- scanner.Text():
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
```

**Step 2: 验证编译**

```bash
go build ./...
```

**Step 3: 提交**

```bash
git add internal/core/singbox/manager.go
git commit -m "feat: add sing-box process manager"
```

---

### Task 3.3: 实现崩溃恢复

**Files:**
- Create: `internal/core/singbox/recovery.go`

**Step 1: 创建恢复模块**

Create `internal/core/singbox/recovery.go`:

```go
package singbox

import (
	"log"
	"os"
	"strconv"
	"syscall"

	"singbox-web/internal/storage"
)

func (m *Manager) RecoverFromCrash() error {
	// Get last status
	lastStatus, err := storage.GetSetting("singbox_status")
	if err != nil || lastStatus != "running" {
		log.Println("No recovery needed, sing-box was not running")
		return nil
	}

	// Get last PID
	pidStr, err := storage.GetSetting("singbox_pid")
	if err != nil {
		return m.restartSingbox()
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return m.restartSingbox()
	}

	// Check if process is still alive
	process, err := os.FindProcess(pid)
	if err != nil {
		return m.restartSingbox()
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		log.Printf("Previous sing-box process (PID %d) not found, restarting", pid)
		return m.restartSingbox()
	}

	log.Printf("Reattaching to existing sing-box process (PID %d)", pid)

	// Reattach to process
	m.mu.Lock()
	m.status = StatusRunning
	m.mu.Unlock()

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "singbox_recover",
		Detail: "Reattached to existing sing-box process",
	})

	return nil
}

func (m *Manager) restartSingbox() error {
	log.Println("Attempting to restart sing-box...")

	// Check if config exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		log.Println("Config file not found, skipping restart")
		storage.SetSetting("singbox_status", "stopped")
		return nil
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "singbox_recover",
		Detail: "Restarting sing-box after crash",
	})

	return m.Start()
}
```

**Step 2: 验证编译**

```bash
go build ./...
```

**Step 3: 提交**

```bash
git add internal/core/singbox/recovery.go
git commit -m "feat: add sing-box crash recovery"
```

---

### Task 3.4: 实现系统 API

**Files:**
- Create: `internal/api/handlers/system.go`

**Step 1: 创建系统处理器**

Create `internal/api/handlers/system.go`:

```go
package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"singbox-web/internal/core/singbox"
	"singbox-web/internal/storage"
)

var singboxManager *singbox.Manager

func InitSystemHandlers(dataDir string) {
	singboxManager = singbox.GetManager(dataDir)
}

type SystemStatus struct {
	SingboxStatus  string `json:"singbox_status"`
	SingboxVersion string `json:"singbox_version,omitempty"`
	SingboxPid     int    `json:"singbox_pid,omitempty"`
	ConfigExists   bool   `json:"config_exists"`
}

func GetSystemStatus(c *gin.Context) {
	status := SystemStatus{
		SingboxStatus: string(singboxManager.GetStatus()),
		SingboxPid:    singboxManager.GetPid(),
	}

	// Check config
	dataDir := c.GetString("dataDir")
	if dataDir == "" {
		dataDir = "data"
	}
	if _, err := os.Stat(dataDir + "/config.json"); err == nil {
		status.ConfigExists = true
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
	version := "0.1.0" // TODO: read from build info

	singboxVersion, err := singbox.GetLatestVersion()
	if err != nil {
		singboxVersion = "unknown"
	}

	c.JSON(http.StatusOK, gin.H{
		"version":         version,
		"singbox_version": singboxVersion,
	})
}

func UpgradeSingbox(c *gin.Context) {
	dataDir := c.GetString("dataDir")
	if dataDir == "" {
		dataDir = "data"
	}

	version, err := singbox.DownloadLatest(dataDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log operation
	storage.DB.Create(&storage.OperationLog{
		Action: "singbox_upgrade",
		Detail: "Upgraded to " + version,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "sing-box upgraded",
		"version": version,
	})
}

func GetGeneratedConfig(c *gin.Context) {
	dataDir := c.GetString("dataDir")
	if dataDir == "" {
		dataDir = "data"
	}

	configPath := dataDir + "/config.json"
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}
```

**Step 2: 更新路由**

Update `internal/api/router.go`, add after auth routes in protected group:

```go
// System routes
system := protected.Group("/system")
{
    system.GET("/status", handlers.GetSystemStatus)
    system.POST("/start", handlers.StartSingbox)
    system.POST("/stop", handlers.StopSingbox)
    system.POST("/restart", handlers.RestartSingbox)
    system.GET("/version", handlers.GetSystemVersion)
    system.POST("/upgrade", handlers.UpgradeSingbox)
    system.GET("/config", handlers.GetGeneratedConfig)
}
```

**Step 3: 更新 main.go 初始化**

Add in `cmd/singbox-web/main.go` after database init:

```go
// Initialize handlers
handlers.InitSystemHandlers(dataDir)

// Recover from crash
manager := singbox.GetManager(dataDir)
if err := manager.RecoverFromCrash(); err != nil {
    log.Printf("Failed to recover: %v", err)
}
```

Add import:

```go
import (
    "singbox-web/internal/api/handlers"
    "singbox-web/internal/core/singbox"
)
```

**Step 4: 验证编译**

```bash
go build ./...
```

**Step 5: 提交**

```bash
git add internal/api/handlers/system.go internal/api/router.go cmd/singbox-web/main.go
git commit -m "feat: add system API for sing-box management"
```

---

## 后续里程碑概要

由于篇幅限制，以下里程碑将在后续计划文档中详细展开：

### M4: 入站管理
- Task 4.1: 入站 CRUD API
- Task 4.2: 入站前端页面

### M5: 出站/订阅管理
- Task 5.1: 订阅 CRUD API
- Task 5.2: 订阅解析器（Clash/V2Ray/Base64）
- Task 5.3: 出站节点 CRUD API
- Task 5.4: 节点测速功能
- Task 5.5: 出站前端页面

### M6: 规则管理
- Task 6.1: 规则集 CRUD API
- Task 6.2: 规则中心预置数据
- Task 6.3: 路由规则 CRUD API
- Task 6.4: 规则前端页面

### M7: 配置生成
- Task 7.1: 配置生成器核心
- Task 7.2: 配置备份机制
- Task 7.3: 应用配置流程

### M8: 定时任务
- Task 8.1: 定时调度器
- Task 8.2: 订阅自动更新
- Task 8.3: 规则集自动更新

### M9: 日志系统
- Task 9.1: 操作日志 API
- Task 9.2: WebSocket 实时日志
- Task 9.3: 日志前端页面

### M10: 集成构建
- Task 10.1: 前端嵌入后端
- Task 10.2: 构建脚本
- Task 10.3: 安装脚本

---

**完成 M1-M3 后，系统将具备：**
- 基础项目结构
- 用户登录认证
- sing-box 下载和进程管理
- 崩溃恢复机制

这是一个可运行的最小系统，可以验证核心架构后再继续实现业务功能。
