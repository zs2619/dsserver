# DS Server

[腾讯 DS 开发组件集](https://gcloud.tencent.com/pages/documents/details.html?projectId=205) 的开源实现（Golang + gRPC）。

## 功能概览

1. 管理 DS Agent 与 DS 进程生命周期  
2. 实时分配 DS（匹配 → Realm → CreateDS）  
3. 快速拉起 UE DS 实例  

当前仓库已实现的核心链路：

```text
Client --TCP(kissnet)--> GameServer --gRPC--> DSC --gRPC--> DSA --进程--> DS
                              ^                    ^
                              |   Redis 集群(控制面) |
                              +------ Mongo(用户数据，GS)
```

| 服务 | 目录 | 职责 |
|------|------|------|
| GameServer (GS) | `gameserver/` | 玩家连接、组队匹配、订阅 Realm ready/fail |
| DSC | `dscserver/` | Agent 注册/心跳、CreateRealm、分配 DS |
| DSA | `dsaserver/` | 本机拉起 DS、上报 CreateOK/DEAD 等事件 |

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

`docker-compose.yaml`：GS / DSC / DSA，**加入外部网络 `ds-infra`**，与 infra 互通。

```bash
make infra                         # 必须先起依赖
docker compose up -d --build       # 再起业务
```

启动顺序：先 `infra`（创建网络与数据服务），再业务 compose。

## 开发环境（Dev Container）

配置在 `.devcontainer/`：

- 合并加载 `docker-compose.infra.yaml` + `docker-compose.extend.yml`  
- 开发容器 `dsservices` 与 Redis/Mongo 同处 `ds-infra`  
- 调试环境变量见 `.vscode/launch.json`  

步骤：

1. 安装 VS Code 扩展 **Dev Containers**  
2. 打开本仓库根目录  
3. **Reopen in Container**  

容器内已注入 `DS_REDISURI` / `DS_MONGOURI` 等，可直接 F5 调试 GS / DSC / DSA。

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
| `make buildimage` | 构建 GS/DSC/DSA 运行镜像 |

## 配置

- 游戏/超时等：`assets/config.json`（由 `DS_CONFIGPATH` 指定目录前缀）  
- 关键变量示例：`DS_GS_PORT`、`DS_DSC_PORT`、`DS_DSA_*`、`DS_DSC_GRPC_URI`、`DS_DSA_AgentID`  

## 许可证

见 `LICENSE`（MIT）。
