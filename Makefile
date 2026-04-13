.PHONY: help dev dev-frontend dev-backend infra infra-down build-frontend build-backend clean

# DouTok 开发命令集
help:
	@echo "DouTok Development Commands"
	@echo "=========================="
	@echo "  make infra          - 启动基础设施 (MySQL, Redis, Kafka, MinIO, SRS)"
	@echo "  make infra-down     - 停止基础设施"
	@echo "  make dev-backend    - 启动后端 (Go gateway)"
	@echo "  make dev-frontend   - 启动前端开发服务器"
	@echo "  make dev            - 启动全部 (infra + backend + frontend)"
	@echo "  make build-backend  - 编译后端二进制"
	@echo "  make build-frontend - 编译前端静态文件"
	@echo "  make build-apk      - 编译 Android APK"
	@echo "  make clean          - 清理构建产物"
	@echo "  make migrate        - 执行数据库迁移"

infra:
	cd deploy/docker && docker compose up -d
	@echo "Waiting for MySQL to be ready..."
	@sleep 10
	@echo "Infrastructure started!"

infra-down:
	cd deploy/docker && docker compose down

dev-backend:
	cd backend && go run ./cmd/gateway

dev-frontend:
	cd frontend && npm run dev

dev: infra
	@echo "Starting backend and frontend..."
	@make dev-backend &
	@make dev-frontend &
	@wait

build-backend:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../dist/gateway ./cmd/gateway

build-frontend:
	cd frontend && npm run build

build-apk: build-frontend
	cd frontend && npx cap sync android && cd android && ./gradlew assembleDebug
	@echo "APK: frontend/android/app/build/outputs/apk/debug/app-debug.apk"

migrate:
	@echo "Running migrations..."
	docker exec -i doutok-mysql mysql -udoutok -pdoutok123 doutok < backend/migration/001_init.sql
	@echo "Migration complete!"

clean:
	rm -rf dist/ frontend/dist/ backend/gateway
