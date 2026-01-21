# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

singbox.arrow.web2 是一个 sing-box 的 Web 管理客户端，用于中转服务器/路由器场景的流量二次分流。

**技术栈:**
- 后端: Go 1.25+ / Gin / GORM / SQLite
- 前端: Vue 3 / TypeScript / Vite / Element Plus / Pinia
- 认证: JWT
- 部署: 单二进制（前端嵌入）

## 常用命令

### 后端
```bash
# 开发模式运行（使用当前目录的 data/）
DEV=1 go run cmd/singbox.arrow.web2/main.go

# 构建
go build -o singbox.arrow.web2 ./cmd/singbox.arrow.web2/

# 运行（使用可执行文件同目录的 data/）
./singbox.arrow.web2
```

### 前端
```bash
cd frontend

# 安装依赖
npm install

# 开发模式（端口 5173，代理 API 到 60017）
npm run dev

# 构建生产版本
npm run build
```

### 服务端口
- 后端 API: `http://localhost:60017`
- 前端开发: `http://localhost:5173`
- 默认账号: `admin` / `123`

## 代码架构

```
├── cmd/singbox.arrow.web2/main.go   # 入口：初始化数据库、恢复崩溃、启动 Gin
├── internal/
│   ├── api/
│   │   ├── router.go                # Gin 路由配置
│   │   ├── middleware/auth.go       # JWT 认证中间件
│   │   └── handlers/                # API 处理器
│   ├── core/singbox/
│   │   ├── downloader.go            # 从 GitHub 下载 sing-box（支持代理）
│   │   ├── manager.go               # 进程管理（启动/停止/重启）
│   │   └── recovery.go              # 崩溃恢复（PID 检测）
│   └── storage/
│       ├── database.go              # SQLite 初始化、设置读写
│       └── models.go                # GORM 模型（9张表）
└── frontend/src/
    ├── api/                         # Axios 客户端
    ├── stores/                      # Pinia 状态管理
    ├── layouts/MainLayout.vue       # 主布局（侧边栏+顶栏）
    ├── views/                       # 页面组件
    └── router/                      # Vue Router 配置
```

## 实现进度

详见 `docs/plans/2026-01-21-singbox-web-implementation.md`

**已完成 (M1-M3):**
- M1: 项目初始化（Go + Vue 骨架）
- M2: 认证系统（JWT 登录/登出/改密）
- M3: sing-box 管理（下载/启停/崩溃恢复）
- 界面中文化 + 主布局

**待实现:**
- M4: 入站管理
- M5: 订阅/出站管理
- M6: 规则集管理
- M7: 配置生成
- M8: 定时任务
- M9: 日志系统
- M10: 集成构建

## 继续开发指南

新 Session 启动时，告诉 Claude:
```
继续开发 singbox.arrow.web2，
读取 docs/plans/2026-01-21-singbox-web-implementation.md，
当前已完成 M1-M3，下一步是 M4 入站管理
```

或使用:
```
/superpowers:execute-plan docs/plans/2026-01-21-singbox-web-implementation.md
从 M4 开始
```

## 数据库模型

位于 `internal/storage/models.go`:
- `Setting`: 键值配置
- `Inbound`: 入站配置
- `Subscription`: 订阅源
- `Outbound`: 出站节点
- `Ruleset`: 规则集
- `Rule`: 路由规则
- `OperationLog`: 操作日志
- `TrafficStat`: 流量统计
- `RuntimeState`: 运行时状态

## API 路由

- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/logout` - 登出（需认证）
- `PUT /api/v1/auth/password` - 修改密码（需认证）
- `GET /api/v1/system/status` - 系统状态
- `POST /api/v1/system/start|stop|restart` - 控制 sing-box
- `POST /api/v1/system/upgrade` - 下载/更新 sing-box
