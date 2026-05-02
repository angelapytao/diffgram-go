.PHONY: build run-api test test-unit lint migrate-up migrate-status \
        docker-up docker-down docker-build codegen help

COMPOSE_FILE := deployments/docker-compose.yaml
GO           := go
GOFLAGS      := -v

## build: 编译全部服务二进制
build:
	$(GO) build $(GOFLAGS) ./cmd/api ./cmd/processor ./cmd/worker

## run-api: 本地启动 api 服务（需先 docker-up 起基础设施）
run-api:
	$(GO) run ./cmd/api

## test: 运行全部测试（含 integration，需要 Docker）
test:
	$(GO) test ./... -timeout 120s

## test-unit: 仅运行单元测试（跳过 integration）
test-unit:
	$(GO) test ./... -short -timeout 30s

## lint: 运行 golangci-lint
lint:
	golangci-lint run ./...

## migrate-up: 执行全部待执行迁移
migrate-up:
	goose -dir migrations mysql "$(DATABASE_URL)" up

## migrate-status: 查看迁移状态
migrate-status:
	goose -dir migrations mysql "$(DATABASE_URL)" status

## docker-up: 启动全部基础设施容器
docker-up:
	docker compose -f $(COMPOSE_FILE) up -d --build

## docker-down: 停止并移除容器
docker-down:
	docker compose -f $(COMPOSE_FILE) down

## docker-build: 仅构建 docker 镜像
docker-build:
	docker compose -f $(COMPOSE_FILE) build

## codegen: 生成 Kitex RPC 代码
codegen:
	./scripts/codegen.sh

## help: 显示可用命令
help:
	@grep -E '^## ' Makefile | sed 's/## //'

.DEFAULT_GOAL := build