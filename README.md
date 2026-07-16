# ITCFG - 配置中台 + 部署包编排系统

## 项目简介

ITCFG 是一套面向企业级多组件部署场景的配置管理与部署自动化系统。

**核心能力：** 运维/开发人员在配置中台按客户环境录入配置 → 一键导出完整部署包 → 现场人员拿到部署包后执行一条命令完成部署。

**零技术门槛、零出错、全可追溯。**

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                     云端                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐ │
│  │ 配置中台(Web) │  │ 制品仓库      │  │ 部署包编排     │ │
│  │ Go + Vue 3   │  │ Harbor       │  │ 服务           │ │
│  └──────────────┘  └──────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────┘
                          │
                    ┌─────▼─────┐
                    │  部署包    │
                    │  .tar.gz  │
                    └─────┬─────┘
                          │
┌─────────────────────────┼───────────────────────────────┐
│                   客户现场                                │
│                    ┌──────▼──────┐                       │
│                    │ Config Agent │                       │
│                    │ (Go 二进制)  │                       │
│                    └─────────────┘                       │
│              解包 → 校验 → 部署 → 启动                    │
└─────────────────────────────────────────────────────────┘
```

## 项目结构

```
itcfg/
├── server/                  # 配置中台后端 (Go + Gin)
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── handler/         # HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── repository/      # 数据访问
│   │   ├── model/           # 数据模型
│   │   ├── middleware/      # 中间件
│   │   └── template/        # 模板引擎
│   └── templates/           # 组件模板
│       ├── nginx/
│       ├── java-app/
│       └── postgresql/
├── web/                     # 配置中台前端 (Vue 3)
├── agent/                   # Config Agent (Go)
│   ├── cmd/agent/           # 入口
│   └── internal/
│       ├── command/         # 命令行处理
│       ├── deploy/          # 部署逻辑
│       ├── validate/        # 校验逻辑
│       └── config/          # 配置处理
├── docker/                  # 开发环境
├── docs/                    # 文档
│   └── 方案.md              # 项目方案
└── README.md
```

## 快速开始

### 前置要求

- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+ (开发环境可通过 Docker 提供)

### 启动开发环境

```bash
# 1. 启动 PostgreSQL
docker-compose -f docker/docker-compose.dev.yml up -d postgres

# 2. 启动后端
cd server
go run cmd/server/main.go

# 3. 启动前端 (另一终端)
cd web
npm install
npm run dev
```

### 编译 Config Agent

```bash
cd agent
go build -o config-agent cmd/agent/main.go
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 配置中台后端 | Go + Gin + GORM |
| 配置中台前端 | Vue 3 + Element Plus |
| 数据库 | PostgreSQL 15+ |
| 制品仓库 | Harbor |
| Config Agent | Go (单一二进制) |
| 模板引擎 | Go text/template |
| 部署方式 | Docker Compose |

## 组件模板

已支持的组件模板（14 个全部就绪）：

- [x] Nginx
- [x] Java App (Spring Boot)
- [x] PostgreSQL
- [x] Redis
- [x] MinIO
- [x] Kafka
- [x] Etcd
- [x] MongoDB
- [x] OnlyOffice
- [x] Elasticsearch
- [x] Collabora Code
- [x] File Viewer
- [x] FileCodeBox
- [x] Docker Compose

## 实施路线

| 阶段 | 功能 | 状态 |
|------|------|------|
| 一期 MVP | 配置中台 + Agent 最小闭环 + 核心组件模板 | ✅ |
| 二期 | 制品版本关联 + Agent 在线模式 + 版本管理/回滚/差异对比 + 配置克隆 | ✅ |
| 三期 | JWT 认证 + 用户管理 + 加密存储 + 配置校验 + 健康检查 + 钉钉/企微通知 + Swagger | ✅ |

## 文档

- [项目方案](docs/方案.md)
- [API 文档](docs/api.md) (待完善)
- [组件模板开发指南](docs/component-template-guide.md) (待完善)