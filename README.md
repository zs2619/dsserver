#  DS 
[腾讯DS开发组件集]( https://gcloud.tencent.com/pages/documents/details.html?projectId=205)的开源实现
# DS管理主要提供以下功能：
1. 管理DS集群
2. 实时分配DS
3. DS快速拉起
# 使用场景
完全掌控DS集群：管理多版本的UE4 DS集群，跟踪DS完整生命周期，支持单机DS多版本混布，实时分配DS，不停服升级DS，集群伸缩，监控DS集群，自定义报警。不限制云运营商，支持私有化部署，用户可根据各地区情况灵活地选择部署方案。
快速启动开发：精心设计的DS管理流程和UE4侧DS管理接入插件，将常见问题的应对固化到DS管理中，使得开发前期决策变得简单、避坑，可灵活满足整个DS开发周期的需求，帮助业务轻松跨越前后端开发沟通无从下手的问题，使得业务能快速有效地启动项目开发。
# 开发技术
Golang GRPC 
# 开发环境
使用vscode docker

环境配置文件在 .devcontainer目录.

调试配置launch.json

直接在vscode中打开工程目录，安装Remote Containers插件, 然后容器内开发调试。

# 测试环境
使用 
docker-compose.yaml

# 发布
使用 Makefile

make build 压缩包

make buildimage 构建镜像