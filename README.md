<div align="center">

<img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go">
<img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vuedotjs" alt="Vue">
<img src="https://img.shields.io/badge/PostgreSQL-15+-4169E1?style=flat&logo=postgresql" alt="PostgreSQL">
<img src="https://img.shields.io/badge/Docker-24+-2496ED?style=flat&logo=docker" alt="Docker">
<img src="https://img.shields.io/badge/license-MIT-green?style=flat" alt="License">
<img src="https://img.shields.io/badge/status-production%20ready-brightgreen?style=flat" alt="Status">

</div>

<br>

# 🏗️ ITCFG — 企业级配置中台 & 部署编排平台

> 配置即代码 · 一键部署 · 零门槛交付

**ITCFG** 是一套面向企业级多组件部署场景的 **配置管理与部署自动化系统**。运维/开发人员在 Web 中台按客户环境录入配置，**一键导出完整部署包**，现场人员拿到部署包后 **执行一条命令即可完成部署**。

<br>

## ✨ 核心亮点

<div align="center">

| 🎯 零技术门槛 | 🔒 零出错风险 | 📋 全链路可追溯 |
|:---:|:---:|:---:|
| 表单化录入，告别手工改配置 | 类型校验 + 版本快照 + 差异对比 | 每次变更留痕，部署记录可查 |

</div>

<br>

## 🏛️ 系统架构

```
┌──────────────────────────────────────────────────────────────────┐
│                          ☁️  云端                                 │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────────┐  │
│  │   🖥️ 配置中台    │  │   📦 制品仓库    │  │   📐 部署包编排   │  │
│  │  Go + Vue 3     │──│  Harbor         │──│  配置 + 制品打包   │  │
│  │  配置录入·渲染·  │  │  镜像·Jar·Helm  │  │  tar.gz 一键导出  │  │
│  │  版本·克隆·通知  │  │                 │  │                  │  │
│  └────────┬────────┘  └─────────────────┘  └────────┬─────────┘  │
│           │                                          │            │
└───────────┼──────────────────────────────────────────┼────────────┘
            │                  📦 部署包                │
            │               customer-env-v1.0.0.tar.gz │
            └──────────────────┬───────────────────────┘
                               │ 下载 / U 盘拷贝
┌──────────────────────────────┼───────────────────────────────────────┐
│                         🏭 客户现场                                    │
│                              │                                        │
│                    ┌─────────▼──────────┐                             │
│                    │   ⚙️ Config Agent   │                             │
│                    │   Go 单一二进制      │                             │
│                    │   ~15MB 零依赖       │                             │
│                    └─────────┬──────────┘                             │
│                              │                                        │
│               📦 解包 → 🔍 校验 → 🚀 部署 → ✅ 启动                   │
│                              │                                        │
│         ┌────────┬────────┬─┴────┬────────┬────────┐                 │
│         ▼        ▼        ▼      ▼        ▼        ▼                 │
│      ┌─────┐ ┌─────┐ ┌──────┐ ┌─────┐ ┌──────┐ ┌──────┐            │
│      │Nginx│ │Java │ │ PG   │ │Redis│ │MinIO │ │  …  │            │
│      └─────┘ └─────┘ └──────┘ └─────┘ └──────┘ └──────┘            │
│                     14+ 组件一键部署                                  │
└──────────────────────────────────────────────────────────────────────┘
```

<br>

## 🚀 快速开始

### 环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | ≥ 1.22 | 后端 + Agent 编译 |
| Node.js | ≥ 18 | 前端开发 |
| Docker | ≥ 24 | 容器运行时 |
| PostgreSQL | ≥ 15 | 数据持久化 |

### 一键启动开发环境

```bash
# 1. 启动依赖服务
docker compose -f docker/docker-compose.dev.yml up -d

# 2. 启动后端 (端口 8080)
cd server && go run cmd/server/main.go

# 3. 启动前端 (端口 5173)
cd web && npm install && npm run dev
```

### 环境变量

```bash
# 数据库
DATABASE_DSN=host=localhost user=itcfg password=itcfg123 dbname=itcfg port=5432

# 安全
ENCRYPTION_KEY=your-32-byte-master-key    # AES-256-GCM 加密密钥
JWT_SECRET=your-jwt-signing-secret        # JWT 签名密钥

# 管理员
ADMIN_USER=admin                          # 默认: admin
ADMIN_PASS=admin123                       # 默认: admin123
```

### 编译 Config Agent

```bash
cd agent
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o config-agent cmd/agent/main.go
```

<br>

## 📂 项目结构

```
itcfg/
├── server/                         # 🖥️ 配置中台后端 (Go + Gin + GORM)
│   ├── cmd/server/                 #   入口 · 依赖注入
│   ├── internal/
│   │   ├── handler/                #   HTTP 处理器 (30+ API)
│   │   ├── service/                #   业务逻辑层
│   │   ├── repository/             #   数据访问层 (PostgreSQL)
│   │   ├── model/                  #   数据模型 (8 实体)
│   │   ├── middleware/             #   CORS · 日志 · JWT · Agent 认证
│   │   ├── template/               #   Go template 渲染引擎
│   │   ├── auth/                   #   JWT 生成/验证 · bcrypt 密码
│   │   ├── crypto/                 #   AES-256-GCM 加密模块
│   │   ├── validate/               #   配置值校验引擎 (8 种类型)
│   │   └── notify/                 #   钉钉/企微/Webhook 通知
│   ├── templates/                  #   14 个组件配置模板
│   └── docs/                       #   Swagger 文档
│
├── web/                            # 🎨 配置中台前端 (Vue 3 + Element Plus)
│   └── src/
│       ├── views/                  #   14 个页面组件
│       ├── api/                    #   API 层 (Token 管理 · 拦截器)
│       └── router/                 #   路由 · 导航守卫
│
├── agent/                          # ⚙️ Config Agent (Go 单一二进制)
│   └── internal/
│       ├── command/                #   CLI 命令 (8 个子命令)
│       ├── deploy/                 #   部署引擎 (离线+在线)
│       ├── validate/               #   环境校验
│       └── config/                 #   配置管理
│
├── docker/                         # 🐳 Docker Compose 开发环境
├── docs/                           # 📖 方案文档
└── .gitignore
```

<br>

## 🛠️ 技术栈

<table>
<tr>
  <td align="center" width="120"><b>后端框架</b></td>
  <td><img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white"> <img src="https://img.shields.io/badge/Gin-00ADD8?logo=go&logoColor=white"> <img src="https://img.shields.io/badge/GORM-00ADD8?logo=go&logoColor=white"></td>
</tr>
<tr>
  <td align="center"><b>前端框架</b></td>
  <td><img src="https://img.shields.io/badge/Vue_3-4FC08D?logo=vuedotjs&logoColor=white"> <img src="https://img.shields.io/badge/Element_Plus-409EFF?logo=element&logoColor=white"> <img src="https://img.shields.io/badge/TypeScript-3178C6?logo=typescript&logoColor=white"></td>
</tr>
<tr>
  <td align="center"><b>数据库</b></td>
  <td><img src="https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white"> <img src="https://img.shields.io/badge/JSONB-4169E1?logo=postgresql&logoColor=white"></td>
</tr>
<tr>
  <td align="center"><b>安全</b></td>
  <td><img src="https://img.shields.io/badge/JWT-000000?logo=jsonwebtokens&logoColor=white"> <img src="https://img.shields.io/badge/AES_256_GCM-EE0000"> <img src="https://img.shields.io/badge/bcrypt-EE0000"></td>
</tr>
<tr>
  <td align="center"><b>部署</b></td>
  <td><img src="https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white"> <img src="https://img.shields.io/badge/Docker_Compose-2496ED?logo=docker&logoColor=white"></td>
</tr>
<tr>
  <td align="center"><b>文档</b></td>
  <td><img src="https://img.shields.io/badge/Swagger-85EA2D?logo=swagger&logoColor=black"> <img src="https://img.shields.io/badge/OpenAPI-85EA2D?logo=openapiinitiative&logoColor=black"></td>
</tr>
</table>

<br>

## 🧩 组件模板

已支持的 **14 个组件配置模板**，全部开箱即用：

<table>
<tr>
  <td>🌐 <b>Nginx</b></td>
  <td>☕ <b>Java App</b></td>
  <td>🐘 <b>PostgreSQL</b></td>
  <td>🔴 <b>Redis</b></td>
</tr>
<tr>
  <td>🪣 <b>MinIO</b></td>
  <td>📨 <b>Kafka</b></td>
  <td>🔗 <b>Etcd</b></td>
  <td>🍃 <b>MongoDB</b></td>
</tr>
<tr>
  <td>📝 <b>OnlyOffice</b></td>
  <td>🔍 <b>Elasticsearch</b></td>
  <td>🤝 <b>Collabora Code</b></td>
  <td>📁 <b>File Viewer</b></td>
</tr>
<tr>
  <td>📦 <b>FileCodeBox</b></td>
  <td>🐳 <b>Docker Compose</b></td>
  <td colspan="2" align="center"><i>模板化架构 · 新增组件零代码</i></td>
</tr>
</table>

每个模板包含：`manifest.yaml` (元信息) + `variables.yaml` (变量定义) + `files/*.tmpl` (配置模板)

<br>

## 📋 功能矩阵

### 🏗️ 配置管理中台

| 模块 | 功能 | 状态 |
|------|------|:--:|
| 👥 **客户管理** | 客户 CRUD · 环境管理 · 环境密钥 | ✅ |
| 🧩 **组件管理** | 组件定义 · 变量管理 · 分组展示 | ✅ |
| ⚙️ **配置录入** | 表单化录入 · 8 种变量类型 · 实时预览 · 模板渲染 | ✅ |
| 🔄 **配置克隆** | 单环境配置克隆 · 完整环境克隆 (配置+制品) | ✅ |
| 📝 **版本管理** | 配置快照 · 版本历史 · 差异对比 · 一键回滚 | ✅ |
| 📦 **制品关联** | 组件制品版本管理 · Harbor 集成 | ✅ |
| 📤 **部署包导出** | 一键打包 tar.gz · SHA256 校验 · 浏览器下载 | ✅ |
| 📊 **部署记录** | 版本标签 · 状态追踪 · 时间线 | ✅ |

### 🔐 安全与治理

| 模块 | 功能 | 状态 |
|------|------|:--:|
| 🔑 **JWT 认证** | 登录 · Token 刷新 · 路由守卫 | ✅ |
| 👤 **用户管理** | 用户 CRUD · 角色 (admin/user) · 状态管理 | ✅ |
| 🔒 **加密存储** | AES-256-GCM 加密 · 密码字段掩码显示 | ✅ |
| ✅ **配置校验** | string/number/password/bool/select/url/ip/port | ✅ |
| 🔔 **通知系统** | 钉钉机器人 · 企业微信 · 通用 Webhook · 加签验证 | ✅ |

### ⚙️ Config Agent

| 命令 | 功能 | 状态 |
|------|------|:--:|
| `import` | 导入离线部署包 (tar.gz) | ✅ |
| `deploy` | 执行部署 (自动检测离线/在线模式) | ✅ |
| `online` | 一键在线部署 (login→pull→deploy→report) | ✅ |
| `login` | 登录配置中台 | ✅ |
| `pull` | 在线拉取配置 | ✅ |
| `validate` | 环境校验 (Docker/端口/磁盘) | ✅ |
| `rollback` | 回滚到上一版本 | ✅ |
| `status` | 查看部署状态 | ✅ |

### 🏥 运维保障

| 模块 | 功能 | 状态 |
|------|------|:--:|
| 🩺 **健康检查** | `/health` · DB 连接 · 模板目录 · 前端面板 | ✅ |
| 📖 **API 文档** | Swagger UI (`/swagger/index.html`) | ✅ |
| 📈 **仪表盘** | 实时统计 · 客户/组件/部署数据 | ✅ |

<br>

## 🔌 API 概览

```
公开接口 (无需认证):
  POST   /api/v1/auth/login                       用户登录
  POST   /api/v1/agent/auth                       Agent 认证
  GET    /api/v1/agent/envs/:envKey/configs        Agent 拉取配置
  POST   /api/v1/agent/envs/:envKey/deploy-report  Agent 上报部署

管理接口 (JWT 认证):
  GET    /api/v1/dashboard/stats                   仪表盘统计
  CRUD   /api/v1/customers                         客户管理
  CRUD   /api/v1/customers/:id/envs                环境管理
  GET    /api/v1/components                        组件管理
  GET    /api/v1/templates                         模板列表
  CRUD   /api/v1/envs/:envId/configs               配置管理
  POST   /api/v1/envs/:envId/configs/preview       配置预览
  POST   /api/v1/envs/:envId/export                导出部署包
  POST   /api/v1/envs/:envId/configs/clone         配置克隆
  POST   /api/v1/envs/:envId/clone-env             环境完整克隆
  GET    /api/v1/envs/:envId/versions              版本列表
  POST   /api/v1/envs/:envId/versions/snapshot     保存快照
  GET    /api/v1/envs/:envId/versions/diff         版本对比
  POST   /api/v1/envs/:envId/versions/rollback     版本回滚
  CRUD   /api/v1/envs/:envId/artifacts             制品版本
  GET    /api/v1/envs/:envId/deploy-records        部署记录
  CRUD   /api/v1/users                             用户管理
  CRUD   /api/v1/notify-configs                    通知配置
  POST   /api/v1/notify-configs/:id/test           通知测试

Swagger:
  GET    /swagger/index.html                       API 文档
```

<br>

## 🎯 Config Agent 使用指南

```bash
# ─── 离线模式 ───
config-agent import --package ./customer-prod-v1.0.0.tar.gz
config-agent validate --env --health
config-agent deploy
config-agent status

# ─── 在线模式 (分步) ───
config-agent login --server https://itcfg.example.com --env-key prod-abc123
config-agent pull
config-agent deploy

# ─── 在线模式 (一键) ───
config-agent online --server https://itcfg.example.com --env-key prod-abc123

# ─── 回滚 ───
config-agent rollback
```

<br>

## 📖 文档

- [📋 项目方案设计](docs/方案.md)
- [📘 Swagger API 文档](http://localhost:8080/swagger/index.html)
- [🧩 组件模板开发指南](docs/component-template-guide.md) *(待完善)*

<br>

## 📊 项目数据

| 指标 | 数值 |
|------|:----:|
| 后端 API 端点 | 30+ |
| 前端页面 | 14 |
| 组件模板 | 14 |
| Go 模块 | 12 |
| 数据实体 | 8 |
| Agent 命令 | 8 |
| 通知渠道 | 3 |

<br>

<div align="center">

**Made with ❤️ by ITCFG Team**

*配置中台 — 让部署像呼吸一样简单*

</div>