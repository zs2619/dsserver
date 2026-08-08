#!/usr/bin/env bash
# 生成 N 个 DSA 热池 Deployment（dsa-warm-0 .. dsa-warm-N-1），每个一段不重叠的游戏口。
#
# 用法：
#   ./deploy/gen-warm-deployments.sh [N] [PORTS_PER_DEPLOY] [BASE_PORT] [REPLICAS] [DSC_URI] > deploy/all-warm.yaml
#   ./deploy/gen-warm-deployments.sh                  # 默认 N=3, 500/段, base=30000, replicas=3
#
# 产物：N 个 Deployment，口段连续不重叠：
#   dsa-warm-0: 30000-30499
#   dsa-warm-1: 30500-30999
#   dsa-warm-2: 31000-31499
# 同节点多 Pod 段不重叠 → 不抢口；podAntiAffinity 把同 ordinal 副本散到不同节点。
# 安全组放行 [BASE_PORT, BASE_PORT + N*PORTS_PER_DEPLOY - 1]。

set -euo pipefail

N="${1:-3}"
PORTS_PER_DEPLOY="${2:-500}"
BASE_PORT="${3:-30000}"
REPLICAS="${4:-3}"
DSC_URI="${5:-dscserver:10000}"

emit_deployment() {
  local idx="$1"
  local min="$2"
  local max="$3"
  local name="dsa-warm-${idx}"

  cat <<EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${name}
  labels:
    app: dsa-warm
    dsa-warm/ordinal: "${idx}"
spec:
  replicas: ${REPLICAS}
  selector:
    matchLabels:
      app: dsa-warm
      dsa-warm/ordinal: "${idx}"
  template:
    metadata:
      labels:
        app: dsa-warm
        dsa-warm/ordinal: "${idx}"
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchLabels:
                  app: dsa-warm
                  dsa-warm/ordinal: "${idx}"
              topologyKey: kubernetes.io/hostname
      containers:
        - name: dsa
          image: ds-dsa-run:latest
          env:
            - { name: DS_CONFIGPATH, value: /dsa }
            - { name: DS_DSA_PORT, value: "11000" }
            - { name: DSA_TEMPLATE_PORT, value: "11001" }
            - { name: DS_DSA_GRPC_PORT, value: "12000" }
            - { name: DS_DSC_GRPC_URI, value: "${DSC_URI}" }
            # Downward API 须写在引用方之前：K8s 按声明顺序展开 $(VAR)
            - name: POD_NAME
              valueFrom:
                fieldRef: { fieldPath: metadata.name }
            - name: POD_IP
              valueFrom:
                fieldRef: { fieldPath: status.podIP }
            - name: HOST_IP
              valueFrom:
                fieldRef: { fieldPath: status.hostIP }
            - { name: DS_DSA_AgentID, value: "\$(POD_NAME)" }
            - { name: DS_DSA_RUNTIME, value: container }
            - { name: TEMPLATE_READY_REQUIRED, value: "true" }
            - { name: GAME_PORT_MIN, value: "${min}" }
            - { name: GAME_PORT_MAX, value: "${max}" }
            - { name: DS_DSA_GRPC_ADDR, value: "\$(POD_IP):12000" }
            - { name: DS_ADVERTISE_IP, value: "\$(HOST_IP)" }
          ports:
            - { name: ds-tcp, containerPort: 11000 }
            - { name: ds-template, containerPort: 11001 }
            - { name: ds-grpc, containerPort: 12000 }
        - name: dstemplate
          image: ds-dstemplate-run:latest
          env:
            - { name: DSA_TEMPLATE_ADDR, value: "127.0.0.1:11001" }
            - { name: DSA_DS_ADDR, value: "127.0.0.1:11000" }
            - name: HOST_IP
              valueFrom:
                fieldRef: { fieldPath: status.hostIP }
            - { name: DS_ADVERTISE_IP, value: "\$(HOST_IP)" }
            - { name: GAME_PORT_MIN, value: "${min}" }
            - { name: GAME_PORT_MAX, value: "${max}" }
          ports:
EOF
  # 生成 [min, max] 的 containerPort + hostPort 条目
  for port in $(seq "$min" "$max"); do
    echo "            - { name: game-${port}, containerPort: ${port}, hostPort: ${port} }"
  done
}

echo "# DSA 热池：${N} 个 Deployment，每段 ${PORTS_PER_DEPLOY} 口，base=${BASE_PORT}，replicas=${REPLICAS}"
echo "# 由 deploy/gen-warm-deployments.sh 生成。安全组放行 ${BASE_PORT}-$((BASE_PORT + N*PORTS_PER_DEPLOY - 1))。"

for ((i = 0; i < N; i++)); do
  min=$((BASE_PORT + i * PORTS_PER_DEPLOY))
  max=$((min + PORTS_PER_DEPLOY - 1))
  emit_deployment "$i" "$min" "$max"
done
