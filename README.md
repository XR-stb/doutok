# DouTok 🎬

> 面向亿级用户设计的短视频平台学习项目

## 技术栈

- **前端**: React 19 + TypeScript + Vite 8 + Capacitor (Android)
- **后端**: Go 1.25 + Gin + MySQL + Redis + Kafka
- **直播**: SRS 流媒体服务器
- **存储**: MinIO 对象存储
- **监控**: Prometheus + Grafana + ELK

## 快速开始

```bash
# 1. 启动基础设施
make infra

# 2. 执行数据库迁移
make migrate

# 3. 启动后端
make dev-backend

# 4. 启动前端 (新终端)
make dev-frontend
```

访问 http://localhost:3000

## 核心学习点

| 模块 | 技术点 |
|------|--------|
| ID 生成 | Snowflake 分布式 ID |
| Feed 推荐 | 召回→粗排→精排→重排 四阶段管线 |
| 评论排序 | 热度+时间+身份 三因子加权 |
| 直播排行 | Redis ZSET 实时排行榜 |
| 点赞系统 | Redis 计数 + MySQL 对账 |
| 聊天系统 | WebSocket + Kafka 异步投递 |
| 分库分表 | Snowflake ID + Hash 分片 |
| 日志系统 | zap 结构化日志 + ELK 集中管理 |

## 项目结构

```
doutok/
├── api/                # Proto & OpenAPI 定义
├── backend/            # Go 后端
│   ├── cmd/gateway/    # API 网关入口
│   ├── internal/       # 内部实现
│   │   ├── config/     # 配置管理
│   │   ├── handler/    # HTTP Handler
│   │   ├── middleware/  # 中间件
│   │   ├── model/      # 数据模型
│   │   └── pkg/        # 公共包
│   │       ├── algorithm/   # 推荐算法
│   │       ├── auth/        # JWT 鉴权
│   │       ├── cache/       # Redis 封装
│   │       ├── logger/      # 日志
│   │       ├── mq/          # Kafka 封装
│   │       └── snowflake/   # ID 生成
│   └── migration/      # 数据库迁移
├── frontend/           # React 前端
│   ├── src/
│   │   ├── components/ # 组件
│   │   ├── pages/      # 页面
│   │   ├── stores/     # 状态管理 (Zustand)
│   │   ├── services/   # API 层
│   │   └── styles/     # 全局样式
│   └── capacitor.config.ts
├── deploy/             # 部署配置
│   ├── docker/         # Docker Compose
│   ├── nginx/          # Nginx 配置
│   └── prometheus/     # 监控配置
├── docs/               # 文档
│   └── architecture.md # 架构设计
└── Makefile            # 开发命令
```

## Debug 模式

在个人资料页连续点击版本号 **7 次**即可激活 Debug Panel，和正式 App 一样的隐藏入口。

## License

MIT
