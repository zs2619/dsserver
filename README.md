# DS Server

[腾讯 DS 开发组件集](https://gcloud.tencent.com/pages/documents/details.html?projectId=205) 的开源实现（Golang + gRPC）。

## 功能概览

1. 管理 DS Agent 与 DS 实例生命周期（模板 + Fork 多实例）  
2. 实时分配 DS（匹配 → Realm → CreateDS）  
3. 快速 Fork UE DS 实例（热 Pod 模板常驻，按局 fork）

当前仓库已实现的核心链路：

```text
Client --TCP(kissnet)--> GameServer --gRPC--> DSC --gRPC--> DSA --模板Fork--> DS
                              ^                    ^
                              |   Redis 集群(控制面) |
                              +------ Mongo(用户数据，GS)
```

| 服务 | 目录 | 职责 |
|------|------|------|
| GameServer (GS) | `gameserver/` | 玩家连接、组队匹配、订阅 Realm ready/fail |
| DSC | `dscserver/` | Agent 注册/心跳、CreateRealm、分配 DS |
| DSA | `dsaserver/` | 端口分配、请求模板 Fork 实例、上报 CreateOK/DEAD 等事件 |
| dstemplate | `dstemplate/` | 模拟 UE 模板 DS：连 DSA 发 TemplateReady、收 ForkInstance 后 fork 实例并报 CreateOK |

## 技术栈

- Go 1.26、gRPC / Protobuf  
- Redis 集群（控制面状态机、Pub/Sub）  
- MongoDB（用户 / 玩家数据）  
- Docker Compose、VS Code Dev Containers  

## 统一基础设施（infra）

`docker-compose.infra.yaml` 提供共用依赖，固定网络名 **`ds-infra`**：

| 组件 | 镜像 | 说明 |
|------|------|------|
| Redis 集群 | `redis:8.6.2` | 6 节点（3 主 3 从），端口 7001–7006 |
| MongoDB | `mongo:8.0.28` | 端口 27017 |

```bash
# 启动 Redis 集群 + Mongo，并初始化集群
make infra

# 停止
make infra-down
```

容器内连接示例（`store.RedisInit` 要求 **host:port**，不要写 `redis://`）：

```text
DS_REDISURI=redis-node-1:7001,redis-node-2:7002,redis-node-3:7003,redis-node-4:7004,redis-node-5:7005,redis-node-6:7006
DS_MONGOURI=mongodb://mongodb-server:27017
DS_MONGODBNAME=ds
```

> 若本机拉不下 Docker Hub 镜像，可先用国内镜像再 tag，例如：  
> `docker pull docker.m.daocloud.io/library/mongo:8.0.28 && docker tag docker.m.daocloud.io/library/mongo:8.0.28 mongo:8.0.28`

## 业务服务 Compose

`docker-compose.yaml`：GS / DSC / DSA / dstemplate，**加入外部网络 `ds-infra`**，与 infra 互通。

```bash
make infra                         # 必须先起依赖
# Dockerfile_*.run 从仓库根 COPY 二进制，需先本地编译：
go build -o game gameserver/main.go
go build -o dsc dscserver/main.go
go build -o dsa dsaserver/main.go
go build -o dstemplate dstemplate/main.go
docker compose up -d --build       # 再起业务
```

启动顺序：先 `infra`，再业务 compose。`dstemplate` 使用 `network_mode: service:dsaserver`（与 DSA 共网，模拟 sidecar）；DSA 等模板 `TemplateReady` 后再向 DSC 注册。游戏口段默认 `30000-30100`（已映射到宿主机）。

## 部署模型（模板 + Fork）

- **process**：DSA + dstemplate 同机（compose 共网）；CreateOK 报 `dsIP:口`，本机/同网玩家可达。
- **container**：热池 Pod（DSA sidecar + dstemplate 同 Pod），普通 Pod 网络 + `hostPort`（仅游戏口）；N 个 Deployment 各占一段不重叠游戏口，`podAntiAffinity` 把同段副本散到不同节点；CreateOK 报 `status.hostIP:口`，公网玩家直连，无网关。
  - 部署 YAML 与分段脚本：`deploy/dsa-warm-template.yaml`、`deploy/gen-warm-deployments.sh`。
  - 关键 env：`GAME_PORT_MIN/MAX`（本 Deployment 段）、`DS_ADVERTISE_IP`（Downward API `status.hostIP`）、`DS_DSA_GRPC_ADDR`（Downward API `status.podIP`）、`DS_DSA_AgentID=$(POD_NAME)`（每 Pod 唯一）。

## 开发环境（Dev Container）

配置在 `.devcontainer/`：

- 合并加载 `docker-compose.infra.yaml` + `docker-compose.extend.yml`  
- 开发容器 `dsservices` 与 Redis/Mongo 同处 `ds-infra`  
- 调试环境变量见 `.vscode/launch.json`（含 `dsa server` / `dstemplate`）  

步骤：

1. 安装 VS Code 扩展 **Dev Containers**  
2. 打开本仓库根目录  
3. **Reopen in Container**  

容器内已注入 `DS_REDISURI` / `DS_MONGOURI` 等；F5 调试 DSA 时需另起 `dstemplate`（或先起 compose 里的模板）。

## 测试

```bash
# 宿主机跑测（需已 make infra，且本机可解析集群；推荐容器测）
make test

# 推荐：构建测试镜像，挂到 ds-infra 跑
make unit_test
```

测试镜像：`Dockerfile.test`（默认连 infra 服务名）。

## 常用 Makefile

| 目标 | 说明 |
|------|------|
| `make infra` / `make infra-down` | 起停 Redis 集群 + Mongo |
| `make unit_test` | 容器内单测 |
| `make test` | 本机 `go test` |
| `make vet` | `go vet` |
| `make pb` | 重新生成 protobuf |
| `make build` | 编译并打出二进制 tar.gz |
| `make buildimage` | 构建 GS/DSC/DSA/dstemplate 运行镜像 |

## 配置

- 游戏/超时等：`assets/config.json`（由 `DS_CONFIGPATH` 指定目录前缀；compose 内分别为 `/game` `/dsc` `/dsa`）  
- 环境变量示例：`DS_GS_PORT`、`DS_DSC_PORT`、`DS_DSA_*`、`DS_DSC_GRPC_URI`、`DS_DSA_AgentID`  
- 模板/Fork 相关：`DS_DSA_RUNTIME`（process\|container）、`DSA_TEMPLATE_PORT`、`TEMPLATE_READY_REQUIRED`、`GAME_PORT_MIN`/`GAME_PORT_MAX`、`DS_ADVERTISE_IP`（默认 `status.hostIP`）

## 已知限制（联调前）

- compose `--build` 依赖仓库根已有 `game`/`dsc`/`dsa`/`dstemplate` 二进制（与既有 DSA/DSC 相同）；或用 `make buildimage`。
- 节点主网卡非公网时 `status.hostIP` 内网不可达，需另注入公网 IP（v1 不做）。
- 安全组须放行本 Deployment 游戏口段；hostPort 条目多（约 500/段），需验证 iptables 规模。
- dstemplate 为协议 mock，不做 UE 级 bind 换口；真 UE DS 读 `GAME_PORT_MIN/MAX` 自行换口。

## 许可证

见 `LICENSE`（MIT）。
