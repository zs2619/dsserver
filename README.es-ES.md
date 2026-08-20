# DS Server

[Conjunto de componentes de desarrollo Tencent DS](https://gcloud.tencent.com/pages/documents/details.html?projectId=205) como implementación de código abierto (Golang + gRPC).

## Resumen de funcionalidades

1. Gestión del ciclo de vida de DS Agent y procesos DS  
2. Asignación en tiempo real de DS (coincidencia → Realm → CreateDS)  
3. Arranque rápido de instancias UE DS  

Los enlaces principales ya implementados en este repositorio:

```text
Client --TCP(kissnet)--> GameServer --gRPC--> DSC --gRPC--> DSA --进程--> DS
                              ^                    ^
                              |   Redis 集群(控制面) |
                              +------ Mongo(用户数据，GS)
```

| Servicio | Directorio | Responsabilidad |
|----------|------------|-----------------|
| GameServer (GS) | `gameserver/` | Conexión de jugadores, emparejamiento de equipos, suscripción a Realm ready/fail |
| DSC | `dscserver/` | Registro/heartbeat de Agent, CreateRealm, asignación de DS |
| DSA | `dsaserver/` | Arranque local de DS, informes de CreateOK/DEAD, etc. |

## Tecnología

- Go 1.26、gRPC / Protobuf  
- Redis 集群（控制面状态机、Pub/Sub）  
- MongoDB（用户 / 玩家数据）  
- Docker Compose、VS Code Dev Containers  

## Infraestructura unificada (infra)

`docker-compose.infra.yaml` proporciona dependencias compartidas con red fijada **`ds-infra`**：

| Componente | Imagen | Descripción |
|------------|--------|-------------|
| Redis 集群 | `redis:8.6.2` | 6 nodos (3 primarios 3 secundarios), puertos 7001–7006 |
| MongoDB | `mongo:8.0.28` | Puerto 27017 |

```bash
# Iniciar Redis 集群 + Mongo e inicializar el clúster
make infra

# Detener
make infra-down
```

Ejemplo de conexión dentro del contenedor（`store.RedisInit` requiere **host:port**，no `redis://`）：

```text
DS_REDISURI=redis-node-1:7001,redis-node-2:7002,redis-node-3:7003,redis-node-4:7004,redis-node-5:7005,redis-node-6:7006
DS_MONGOURI=mongodb://mongodb-server:27017
DS_MONGODBNAME=ds
```

> Si no puedes descargar imágenes de Docker Hub en tu máquina, primero usa un registro de imagen local y luego etiquétalo，如：  
> `docker pull docker.m.daocloud.io/library/mongo:8.0.28 && docker tag docker.m.daocloud.io/library/mongo:8.0.28 mongo:8.0.28`

## Servicios de negocio Compose

`docker-compose.yaml`：GS / DSC / DSA，**se une a la red externa `ds-infra`**，conectado con infra.

```bash
make infra                         # Primero iniciar dependencias
docker compose up -d --build       # Luego iniciar los servicios
```

Secuencia de inicio：Primero `infra`（crea red y servicios de datos），luego los servicios de negocio compose.

## Entorno de desarrollo（Dev Container）

Configuración en `.devcontainer/`：

- Carga combinada de `docker-compose.infra.yaml` + `docker-compose.extend.yml`  
- Contenedor de desarrollo `dsservices` y Redis/Mongo comparten la red `ds-infra`  
- Variables de entorno de depuración en `.vscode/launch.json`  

Pasos：

1. Instalar extensión VS Code **Dev Containers**  
2. Abrir directorio raíz del repositorio  
3. **Reabrir en Contenedor**  

Dentro del contenedor ya están inyectadas las variables `DS_REDISURI` / `DS_MONGOURI`，puedes depurar directamente GS / DSC / DSA con F5.

## Pruebas

```bash
# Ejecutar pruebas en el host（需要已 make infra，且本机可解析集群；推荐容器测）
make test

# Recomendado：Construir imagen de prueba，挂到 ds-infra 跑
make unit_test
```

Imagen de prueba：`Dockerfile.test`（默认连 infra 服务名）。

## Makefile común

| Objetivo | Descripción |
|----------|-------------|
| `make infra` / `make infra-down` | Iniciar/detener Redis 集群 + Mongo |
| `make unit_test` | Pruebas unitarias en contenedor |
| `make test` | `go test` en el host |
| `make vet` | `go vet` |
| `make pb` | Regenerar protobuf |
| `make build` | Compilar y generar binario tar.gz |
| `make buildimage` | Construir imágenes GS/DSC/DSA |

## Configuración

- Juego/tiempo de espera等：`assets/config.json`（由 `DS_CONFIGPATH` 指定目录前缀）  
- Variables clave示例：`DS_GS_PORT`、`DS_DSC_PORT`、`DS_DSA_*`、`DS_DSC_GRPC_URI`、`DS_DSA_AgentID`  

## Licencia

Ver `LICENSE`（MIT）。
