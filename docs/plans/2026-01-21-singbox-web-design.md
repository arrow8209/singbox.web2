# singbox-web 设计文档

## 概述

singbox-web 是一个 sing-box 的 Web 管理客户端，运行在 x64 Linux 系统上，提供可视化的配置管理、订阅导入、规则集管理等功能。

## 使用场景

项目分三个阶段开发，支持三种使用场景：

| 阶段 | 场景 | 核心功能 |
|------|------|---------|
| 第一阶段 | 中转服务器 | 链式代理，流量中转到不同落地节点 |
| 第二阶段 | 个人本地代理 | 本地科学上网，类似 Clash Verge |
| 第三阶段 | 路由器 | 分流网关，局域网流量分配到不同出口 |

## 默认配置

| 配置项 | 默认值 |
|-------|-------|
| Web 访问端口 | 60017 |
| 初始用户名 | admin |
| 初始密码 | 123 |

## 技术栈

- **后端**: Go
- **前端**: Vue 3
- **数据库**: SQLite
- **架构**: 单体架构，前端编译后嵌入 Go 二进制

## 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    singbox-web 二进制                    │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────────────────┐   │
│  │   Vue 3 前端     │  │         Go 后端             │   │
│  │   (静态文件)     │  │                             │   │
│  │                 │  │  ┌─────────┐ ┌───────────┐  │   │
│  │  - 仪表盘        │  │  │ HTTP API│ │ WebSocket │  │   │
│  │  - 入站管理      │  │  └────┬────┘ └─────┬─────┘  │   │
│  │  - 出站/订阅     │  │       │             │        │   │
│  │  - 规则集中心    │  │  ┌────▼─────────────▼────┐  │   │
│  │  - 计划任务      │  │  │      业务逻辑层        │  │   │
│  │  - 日志查看      │  │  └───────────┬───────────┘  │   │
│  │  - 系统设置      │  │              │              │   │
│  └─────────────────┘  │  ┌───────────▼───────────┐  │   │
│                       │  │ SQLite │ sing-box │ 日志 │  │   │
│                       │  └───────────────────────┘  │   │
│                       └─────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 核心组件

- **HTTP API**: RESTful 接口，处理配置 CRUD、订阅解析、规则管理
- **WebSocket**: 实时推送 sing-box 日志、连接状态、流量统计
- **业务逻辑层**: 订阅解析器、规则集下载、定时调度、配置生成
- **数据层**: SQLite 存储配置，直接管理 sing-box 子进程

### 日志模块

| 日志类型 | 内容 | 存储方式 |
|---------|------|---------|
| sing-box 日志 | 连接日志、DNS 查询、规则匹配 | 实时流式 + 文件轮转 |
| 系统操作日志 | 用户登录、配置变更、订阅更新 | SQLite 存储 |
| 流量统计日志 | 各节点/规则的流量统计 | SQLite 存储 |

## 数据库设计

```sql
-- 系统配置
-- 存储内容包括：
--   web_port: Web 服务端口 (默认 60017)
--   username: 用户名
--   password_hash: 密码哈希
--   jwt_secret: JWT 密钥
--   singbox_path: sing-box 二进制路径
--   download_proxy_enabled: 下载代理开关 (true/false)
--   download_proxy_url: 下载代理地址 (如 http://127.0.0.1:7890 或 socks5://127.0.0.1:1080)
CREATE TABLE settings (
    key         TEXT PRIMARY KEY,
    value       TEXT,
    updated_at  DATETIME
);

-- 入站配置
CREATE TABLE inbounds (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,  -- http/socks/mixed/tproxy/tun/shadowsocks/vmess/trojan...
    config      TEXT NOT NULL,  -- JSON 格式的详细配置
    enabled     BOOLEAN DEFAULT 1,
    created_at  DATETIME,
    updated_at  DATETIME
);

-- 订阅源
CREATE TABLE subscriptions (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    url         TEXT NOT NULL,
    type        TEXT,           -- auto/singbox/clash/v2ray/base64
    update_interval INTEGER,    -- 更新间隔（小时）
    last_update DATETIME,
    enabled     BOOLEAN DEFAULT 1
);

-- 出站节点（从订阅解析或手动添加）
CREATE TABLE outbounds (
    id          INTEGER PRIMARY KEY,
    subscription_id INTEGER,    -- NULL 表示手动添加
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,  -- direct/block/shadowsocks/vmess/trojan/vless/hysteria2...
    server      TEXT,
    port        INTEGER,
    config      TEXT NOT NULL,  -- JSON 完整配置
    latency     INTEGER,        -- 最近测速结果(ms)
    enabled     BOOLEAN DEFAULT 1,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);

-- 规则集
CREATE TABLE rulesets (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,  -- remote/local
    format      TEXT NOT NULL,  -- source/binary
    url         TEXT,
    update_interval INTEGER,
    last_update DATETIME,
    enabled     BOOLEAN DEFAULT 1
);

-- 路由规则
CREATE TABLE rules (
    id          INTEGER PRIMARY KEY,
    priority    INTEGER NOT NULL,
    type        TEXT NOT NULL,  -- domain/ip/geoip/geosite/ruleset/process...
    value       TEXT NOT NULL,
    outbound_tag TEXT NOT NULL, -- 对应的出站标签
    enabled     BOOLEAN DEFAULT 1
);

-- 操作日志
CREATE TABLE operation_logs (
    id          INTEGER PRIMARY KEY,
    action      TEXT NOT NULL,
    detail      TEXT,
    created_at  DATETIME
);

-- 流量统计
CREATE TABLE traffic_stats (
    id          INTEGER PRIMARY KEY,
    target_type TEXT NOT NULL,  -- outbound/inbound/rule
    target_name TEXT NOT NULL,
    upload      INTEGER DEFAULT 0,
    download    INTEGER DEFAULT 0,
    date        DATE NOT NULL
);

-- 运行时状态（用于崩溃恢复）
CREATE TABLE runtime_state (
    key         TEXT PRIMARY KEY,
    value       TEXT,
    updated_at  DATETIME
);
```

## API 设计

```
基础路径: /api/v1
认证方式: JWT Token (登录后获取)

/auth
  POST /login              登录获取 Token
  POST /logout             登出
  PUT  /password           修改密码

/system
  GET    /status           sing-box 运行状态
  POST   /start            启动 sing-box
  POST   /stop             停止 sing-box
  POST   /restart          重启 sing-box
  GET    /config           获取当前生成的完整配置
  GET    /version          获取版本信息
  POST   /upgrade          升级 sing-box 二进制

/inbounds
  GET    /                 入站列表
  POST   /                 创建入站
  PUT    /:id              更新入站
  DELETE /:id              删除入站
  PATCH  /:id/toggle       启用/禁用

/subscriptions
  GET    /                 订阅列表
  POST   /                 添加订阅
  PUT    /:id              更新订阅
  DELETE /:id              删除订阅
  POST   /:id/refresh      手动刷新订阅

/outbounds
  GET    /                 出站节点列表
  POST   /                 手动添加节点
  PUT    /:id              更新节点
  DELETE /:id              删除节点
  POST   /test             批量测速
  POST   /:id/test         单节点测速

/rulesets
  GET    /                 规则集列表
  POST   /                 添加规则集
  PUT    /:id              更新规则集
  DELETE /:id              删除规则集
  POST   /:id/refresh      刷新规则集
  GET    /center           规则集中心（可下载列表）

/rules
  GET    /                 路由规则列表
  POST   /                 添加规则
  PUT    /:id              更新规则
  DELETE /:id              删除规则
  PUT    /reorder          调整规则优先级

/scheduler
  GET    /tasks            计划任务列表
  PUT    /tasks/:type      更新任务配置
  POST   /tasks/:type/run  立即执行

/logs
  GET    /operations       操作日志查询
  GET    /traffic          流量统计查询
  WS     /realtime         实时日志 WebSocket
```

## 前端页面结构

```
┌─────────────────────────────────────────────────────────┐
│  侧边导航                          顶栏：状态 + 用户菜单  │
├──────────────┬──────────────────────────────────────────┤
│              │                                          │
│  仪表盘       │   主内容区域                             │
│              │                                          │
│  入站管理     │                                          │
│              │                                          │
│  出站管理     │                                          │
│    └ 订阅源   │                                          │
│    └ 节点列表 │                                          │
│              │                                          │
│  规则配置     │                                          │
│    └ 路由规则 │                                          │
│    └ 规则集   │                                          │
│              │                                          │
│  计划任务     │                                          │
│              │                                          │
│  日志        │                                          │
│    └ 实时日志 │                                          │
│    └ 操作记录 │                                          │
│    └ 流量统计 │                                          │
│              │                                          │
│  系统设置     │                                          │
│              │                                          │
└──────────────┴──────────────────────────────────────────┘
```

### 各页面功能

| 页面 | 核心功能 |
|------|---------|
| 仪表盘 | sing-box 运行状态、快捷启停、实时流量图表、节点延迟概览 |
| 入站管理 | 添加/编辑入站、协议配置、端口管理、启用/禁用 |
| 订阅源 | 添加/编辑订阅 URL、手动刷新、查看解析状态 |
| 节点列表 | 所有出站节点、手动添加/编辑节点、批量测速、启用/禁用 |
| 路由规则 | 规则列表、拖拽排序、快速添加常用规则 |
| 规则集 | 已添加的规则集、更新状态、规则中心下载、URL添加、文件上传 |
| 计划任务 | 订阅/规则集更新周期设置 |
| 实时日志 | WebSocket 推送、级别过滤、关键字搜索 |
| 操作记录 | 系统操作历史、分页查询 |
| 流量统计 | 按节点/规则统计、图表展示 |
| 系统设置 | 账户密码、sing-box 路径、自动下载/升级、下载代理配置 |

### 规则集页面设计

```
┌─────────────────────────────────────────────────────────┐
│  规则集                                    [+ 添加规则集] │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  已添加的规则集列表                                       │
│  ┌─────────────────────────────────────────────────┐   │
│  │ ☑ AdGuard DNS Filter    remote  2小时前更新  🔄 🗑 │   │
│  │ ☑ China IP List         remote  1天前更新    🔄 🗑 │   │
│  │ ☑ my-custom-rules       local   手动上传     ✏️ 🗑 │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────┘

点击 [+ 添加规则集] 弹窗，支持三种来源：
1. 规则中心 - 从预置的规则源分类浏览、搜索、一键添加
2. 输入 URL - 手动填写远程规则集地址
3. 上传文件 - 上传本地 .srs / .json 规则文件
```

## 核心业务逻辑

### 订阅解析流程

```
订阅 URL 输入
     ↓
自动检测格式（或用户指定）
     ↓
┌─────────────┬─────────────┬─────────────┬─────────────┐
│  sing-box   │    Clash    │   V2Ray     │   Base64    │
│   JSON      │    YAML     │   JSON      │   链接列表   │
└──────┬──────┴──────┬──────┴──────┬──────┴──────┬──────┘
       │             │             │             │
       └─────────────┴─────────────┴─────────────┘
                          ↓
              统一转换为内部节点格式
                          ↓
              存入 outbounds 表（关联 subscription_id）
```

支持的节点协议解析：
- Shadowsocks (ss://)
- VMess (vmess://)
- VLESS (vless://)
- Trojan (trojan://)
- Hysteria2 (hysteria2://)
- TUIC (tuic://)
- WireGuard
- Direct / Block

### 配置生成流程

```
用户点击「启动」或「应用配置」
              ↓
┌─────────────────────────────────────────┐
│           配置生成器                      │
├─────────────────────────────────────────┤
│  1. 读取 inbounds 表 → 生成入站配置       │
│  2. 读取 outbounds 表 → 生成出站配置      │
│  3. 读取 rules 表 → 生成路由规则          │
│  4. 读取 rulesets 表 → 引用规则集         │
│  5. 合并系统设置 (DNS/日志/实验性功能)     │
└─────────────────────────────────────────┘
              ↓
      生成完整 sing-box JSON 配置
              ↓
      写入 config.json 文件
              ↓
      调用 sing-box run -c config.json
```

### sing-box 进程管理

```go
type SingBoxManager struct {
    process    *os.Process
    configPath string
    status     string  // stopped / running / error
}

func (m *SingBoxManager) Start()   // 启动进程，捕获 stdout/stderr
func (m *SingBoxManager) Stop()    // 优雅停止 (SIGTERM)
func (m *SingBoxManager) Restart() // Stop + Start
func (m *SingBoxManager) Reload()  // 重新生成配置并重启
func (m *SingBoxManager) Logs()    // 返回日志 channel，供 WebSocket 推送
```

### 现场恢复机制

```
┌─────────────────────────────────────────────────────────┐
│                    Web 服务启动流程                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   启动 singbox-web                                      │
│         ↓                                               │
│   读取 runtime_state 表                                  │
│         ↓                                               │
│   ┌─────────────────────────────────────┐               │
│   │ 上次 sing-box 状态 == running ?     │               │
│   └──────────────┬──────────────────────┘               │
│          是 ↓              ↓ 否                         │
│   ┌──────────────────┐  ┌──────────────────┐           │
│   │ 检查进程是否存活   │  │ 保持停止状态     │           │
│   └────────┬─────────┘  └──────────────────┘           │
│    存活 ↓      ↓ 不存活                                  │
│   ┌────────┐  ┌────────────────────┐                   │
│   │ 重新接管 │  │ 使用上次配置重新启动 │                   │
│   └────────┘  └────────────────────┘                   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**runtime_state 表存储内容：**
- singbox_status: running / stopped
- singbox_pid: 进程 ID
- last_config_hash: 上次配置的哈希值
- last_start_time: 上次启动时间

**恢复策略：**

| 场景 | 处理方式 |
|------|---------|
| Web 崩溃，sing-box 还在运行 | 通过 PID 重新接管进程 |
| Web 崩溃，sing-box 也挂了 | 用上次的配置自动重启 sing-box |
| 配置文件损坏 | 从数据库重新生成配置 |
| 数据库损坏 | 保留最近 N 份配置备份，可手动恢复 |

**配置备份机制：**
- 每次生成配置前，备份当前配置到 `backups/config_YYYYMMDD_HHMMSS.json`
- 保留最近 10 份备份，自动清理旧备份
- 前端提供「恢复历史配置」功能

## 下载代理配置

下载 sing-box 二进制、订阅、规则集等网络资源时，可能需要通过代理访问。系统提供统一的下载代理配置。

### 配置项

| 配置项 | 说明 | 示例 |
|-------|------|------|
| download_proxy_enabled | 是否启用下载代理 | true / false |
| download_proxy_url | 代理服务器地址 | http://127.0.0.1:7890 |

### 支持的代理协议

- HTTP 代理：`http://host:port`
- HTTPS 代理：`https://host:port`
- SOCKS5 代理：`socks5://host:port`

### 应用场景

以下操作会使用下载代理（如果启用）：
1. 自动下载/升级 sing-box 二进制
2. 刷新订阅（拉取订阅 URL）
3. 下载/更新规则集
4. 从规则中心下载规则集

### 前端界面

在「系统设置」页面提供：
- 下载代理开关（启用/禁用）
- 代理地址输入框
- 测试连接按钮（验证代理可用性）

## 定时任务设计

```
┌─────────────────────────────────────────────────────────┐
│                    定时任务调度器                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  任务类型          可选间隔              默认值           │
│  ─────────────────────────────────────────────────       │
│  订阅更新          1h / 6h / 12h / 24h / 7d    24h       │
│  规则集更新        1h / 6h / 12h / 24h / 7d    24h       │
│                                                         │
│  执行流程：                                              │
│  1. 到达执行时间                                         │
│  2. 拉取远程数据                                         │
│  3. 对比是否有变更（无变更则跳过）                         │
│  4. 更新数据库                                           │
│  5. 根据配置决定是否自动重载 sing-box                     │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 项目目录结构

```
singbox-web/
├── cmd/
│   └── singbox-web/
│       └── main.go              # 程序入口
│
├── internal/
│   ├── api/                     # HTTP API 层
│   │   ├── router.go
│   │   ├── middleware/
│   │   │   └── auth.go
│   │   └── handlers/
│   │       ├── auth.go
│   │       ├── system.go
│   │       ├── inbound.go
│   │       ├── subscription.go
│   │       ├── outbound.go
│   │       ├── ruleset.go
│   │       ├── rule.go
│   │       ├── scheduler.go
│   │       └── log.go
│   │
│   ├── core/                    # 核心业务逻辑
│   │   ├── singbox/             # sing-box 进程管理
│   │   │   ├── manager.go
│   │   │   └── recovery.go
│   │   ├── parser/              # 订阅解析器
│   │   │   ├── parser.go
│   │   │   ├── clash.go
│   │   │   ├── v2ray.go
│   │   │   ├── singbox.go
│   │   │   └── base64.go
│   │   ├── generator/           # 配置生成器
│   │   │   └── config.go
│   │   └── scheduler/           # 定时任务
│   │       └── scheduler.go
│   │
│   ├── storage/                 # 数据层
│   │   ├── database.go
│   │   ├── models.go
│   │   └── migrations/
│   │
│   └── websocket/               # WebSocket 日志推送
│       └── hub.go
│
├── frontend/                    # Vue 3 前端项目
│   ├── src/
│   │   ├── views/
│   │   ├── components/
│   │   ├── api/
│   │   ├── stores/
│   │   └── router/
│   ├── package.json
│   └── vite.config.js
│
├── data/                        # 运行时数据（.gitignore）
│   ├── singbox-web.db           # SQLite 数据库
│   ├── config.json              # 生成的 sing-box 配置
│   ├── backups/                 # 配置备份
│   ├── rulesets/                # 下载的规则集文件
│   └── logs/                    # 日志文件
│
├── configs/
│   └── ruleset-center.json      # 规则中心预置数据源
│
├── scripts/
│   ├── build.sh                 # 构建脚本
│   └── install.sh               # 安装脚本
│
├── go.mod
├── go.sum
└── README.md
```

## 第一阶段 MVP 范围（中转服务器）

### 必须实现

**系统核心：**
- 单用户登录认证（初始账号 admin/123）
- sing-box 自动下载
- sing-box 进程管理（启动/停止/重启）
- 崩溃恢复机制
- 配置备份
- 下载代理配置（支持 HTTP/HTTPS/SOCKS5 代理）

**入站管理：**
- Shadowsocks / VMess / Trojan / VLESS 入站
- Hysteria2 / TUIC 入站

**出站管理：**
- 订阅导入（sing-box / Clash / V2Ray / Base64）
- 手动添加节点
- 节点测速
- 节点启用/禁用

**规则配置：**
- 路由规则管理
- 规则集管理（添加/更新/删除）
- 规则中心（预置常用规则源）

**定时任务：**
- 订阅定时更新
- 规则集定时更新

**日志与监控：**
- 实时日志查看
- 操作日志记录

**界面：**
- 仪表盘（状态概览、快捷操作）
- 所有管理页面

### 可延后

**第二阶段（个人本地代理）：**
- HTTP / SOCKS5 / Mixed 入站
- TUN 入站（系统全局代理）

**第三阶段（路由器）：**
- TProxy 入站（透明代理）
- 局域网设备管理

**其他：**
- 流量统计图表
- 深色模式

## 参考项目

- https://github.com/GUI-for-Cores/GUI.for.SingBox
- https://github.com/alireza0/s-ui
