# MyLinear 构建脚本
# 使用方式：make <command>

.PHONY: dev build test down clean help

# 默认目标
.DEFAULT_GOAL := help

# 颜色输出
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

# ============================================================================
# 开发环境
# ============================================================================

# 启动完整开发环境（容器 + 后端 + 前端，不含 Caddy）
dev: infra-up backend-dev frontend-dev

# 启动基础设施容器（PostgreSQL、Redis、MinIO）
infra-up:
	@echo "$(GREEN)启动基础设施容器...$(NC)"
	docker compose up -d postgres redis minio
	@echo "$(GREEN)等待服务就绪...$(NC)"
	@sleep 2

# 停止基础设施容器
infra-down:
	@echo "$(YELLOW)停止基础设施容器...$(NC)"
	docker compose down

# 启动后端开发服务器
backend-dev:
	@echo "$(GREEN)启动后端服务...$(NC)"
	cd server && go run ./cmd/server

# 启动前端开发服务器
frontend-dev:
	@echo "$(GREEN)启动前端服务...$(NC)"
	cd web && npm run dev

# ============================================================================
# 构建
# ============================================================================

# 构建后端二进制
build:
	@echo "$(GREEN)构建后端...$(NC)"
	cd server && go build -o bin/server ./cmd/server

# 构建前端产物
build-frontend:
	@echo "$(GREEN)构建前端...$(NC)"
	cd web && npm run build

# 构建生产镜像
build-docker:
	@echo "$(GREEN)构建 Docker 镜像...$(NC)"
	docker compose build

# ============================================================================
# 测试
# ============================================================================

# 运行所有测试
test: test-backend test-frontend

# 运行后端测试
test-backend:
	@echo "$(GREEN)运行后端测试...$(NC)"
	cd server && go test -v ./...

# 运行前端测试
test-frontend:
	@echo "$(GREEN)运行前端测试...$(NC)"
	cd web && npm test

# ============================================================================
# 清理
# ============================================================================

# 停止并清理所有容器
down:
	@echo "$(YELLOW)停止所有服务...$(NC)"
	docker compose down

# 清理构建产物和容器数据
clean: down
	@echo "$(YELLOW)清理构建产物...$(NC)"
	rm -rf server/bin/
	rm -rf web/dist/
	docker compose down -v

# ============================================================================
# 实用工具
# ============================================================================

# 格式化代码
fmt:
	@echo "$(GREEN)格式化 Go 代码...$(NC)"
	cd server && go fmt ./...
	@echo "$(GREEN)格式化前端代码...$(NC)"
	cd web && npm run lint -- --fix || true

# 检查代码质量
lint:
	@echo "$(GREEN)检查后端代码...$(NC)"
	cd server && go vet ./...
	@echo "$(GREEN)检查前端代码...$(NC)"
	cd web && npm run lint || true

# 查看容器日志
logs:
	docker compose logs -f

# 查看容器状态
ps:
	docker compose ps

# ============================================================================
# 帮助
# ============================================================================

help:
	@echo "MyLinear 构建命令"
	@echo ""
	@echo "$(YELLOW)开发环境:$(NC)"
	@echo "  make dev          启动完整开发环境（容器 + 后端 + 前端）"
	@echo "  make infra-up     启动基础设施容器"
	@echo "  make infra-down   停止基础设施容器"
	@echo "  make backend-dev  启动后端开发服务器"
	@echo "  make frontend-dev 启动前端开发服务器"
	@echo ""
	@echo "$(YELLOW)构建:$(NC)"
	@echo "  make build        构建后端二进制"
	@echo "  make build-frontend 构建前端产物"
	@echo "  make build-docker 构建 Docker 镜像"
	@echo ""
	@echo "$(YELLOW)测试:$(NC)"
	@echo "  make test         运行所有测试"
	@echo "  make test-backend 运行后端测试"
	@echo "  make test-frontend 运行前端测试"
	@echo ""
	@echo "$(YELLOW)清理:$(NC)"
	@echo "  make down         停止所有容器"
	@echo "  make clean        清理所有构建产物和容器数据"
	@echo ""
	@echo "$(YELLOW)工具:$(NC)"
	@echo "  make fmt          格式化代码"
	@echo "  make lint         检查代码质量"
	@echo "  make logs         查看容器日志"
	@echo "  make ps           查看容器状态"
