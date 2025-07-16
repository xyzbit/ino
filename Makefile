# Makefile for ino Knowledge System

# 基础变量
APP_NAME := ino
BINARY_NAME := server
GO_VERSION := 1.21.2
BUILD_DIR := bin
DOCKER_IMAGE := ino-system
DOCKER_TAG := latest

# Go 编译参数
CGO_ENABLED := 0
# Docker构建使用Linux
DOCKER_GOOS := linux
DOCKER_GOARCH := amd64
# 本地构建使用当前系统
LOCAL_GOOS := $(shell go env GOOS)
LOCAL_GOARCH := $(shell go env GOARCH)

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "ino Knowledge System - 构建和部署工具"
	@echo ""
	@echo "使用方法:"
	@echo "  make [target]"
	@echo ""
	@echo "目标:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 清理构建文件
.PHONY: clean
clean: ## 清理构建文件和缓存
	@echo "🧹 清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@go clean -cache
	@docker system prune -f
	@echo "✅ 清理完成"

# 下载依赖
.PHONY: deps
deps: ## 下载Go依赖
	@echo "📦 下载依赖包..."
	@go mod download
	@go mod tidy
	@echo "✅ 依赖下载完成"

# 构建本地二进制文件
.PHONY: build
build: deps ## 构建本地运行的应用程序
	@echo "🔨 构建本地应用程序..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(LOCAL_GOOS) GOARCH=$(LOCAL_GOARCH) \
		go build -a -installsuffix cgo \
		-ldflags "-X main.version=$(shell git describe --tags --always --dirty)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		cmd/main.go
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建Docker二进制文件
.PHONY: build-docker
build-docker: deps ## 构建Docker镜像用的应用程序
	@echo "🔨 构建Docker应用程序..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(DOCKER_GOOS) GOARCH=$(DOCKER_GOARCH) \
		go build -a -installsuffix cgo \
		-ldflags "-X main.version=$(shell git describe --tags --always --dirty)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		cmd/main.go
	@echo "✅ Docker构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 本地运行
.PHONY: run
run: build ## 本地运行应用程序
	@echo "🚀 启动应用程序..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 开发模式运行
.PHONY: dev
dev: ## 开发模式运行（不构建二进制文件）
	@echo "🔧 开发模式启动..."
	@go run cmd/main.go

# 构建Docker镜像
.PHONY: docker-build
docker-build: build-docker ## 构建Docker镜像
	@echo "🐳 构建Docker镜像..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# 启动所有服务（不包括API）
.PHONY: services-up
services-up: ## 启动基础设施服务（MySQL, Redis, Milvus, Neo4j, Attu）
	@echo "🚀 启动基础设施服务..."
	@docker-compose up -d mysql redis etcd minio milvus-standalone neo4j attu
	@echo "✅ 基础设施服务启动完成"
	@echo "🌐 Attu 管理界面: http://localhost:3000"
	@echo "等待服务初始化..."
	@sleep 10
	@docker-compose ps

# 启动所有服务
.PHONY: up
up: docker-build ## 构建并启动所有服务
	@echo "🚀 启动所有服务..."
	@docker-compose up -d
	@echo "✅ 所有服务启动完成"
	@echo "🌐 应用地址: http://localhost:8080"
	@echo "📊 健康检查: http://localhost:8080/health"
	@sleep 5
	@docker-compose ps

# 停止所有服务
.PHONY: down
down: ## 停止所有服务
	@echo "🛑 停止所有服务..."
	@docker-compose down
	@echo "✅ 服务停止完成"

# 重启所有服务
.PHONY: restart
restart: down up ## 重启所有服务

# 查看服务日志
.PHONY: logs
logs: ## 查看所有服务日志
	@docker-compose logs -f

# 查看API服务日志
.PHONY: logs-api
logs-api: ## 查看API服务日志
	@docker-compose logs -f ino-api

# 进入API容器
.PHONY: shell
shell: ## 进入API容器shell
	@docker-compose exec ino-api sh

# 健康检查
.PHONY: health
health: ## 检查服务健康状态
	@echo "🔍 检查服务健康状态..."
	@curl -f http://localhost:8080/health || echo "❌ API服务未响应"
	@docker-compose ps

# 查看服务状态
.PHONY: status
status: ## 查看服务状态
	@docker-compose ps

# 测试
.PHONY: test
test: ## 运行测试
	@echo "🧪 运行测试..."
	@go test -v ./...

# 安全检查
.PHONY: security
security: ## 运行安全检查
	@echo "🔒 运行安全检查..."
	@command -v gosec >/dev/null 2>&1 || { echo "请安装gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; exit 1; }
	@gosec ./...

# 性能测试
.PHONY: bench
bench: ## 运行性能测试
	@echo "📈 运行性能测试..."
	@go test -bench=. -benchmem ./...

# 完整的CI流程
.PHONY: ci
ci: clean deps fmt test security build ## 完整的CI流程