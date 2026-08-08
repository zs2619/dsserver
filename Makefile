Version=v1
BACKEND_APP_NAME=ds
BACKEND_SERVERJOB_NAME=ds-job

DOCKER_BUILD_IMAGE_TAG=ds-build

DOCKER_RUN_GAME_IMAGE_TAG=ds-game-run
DOCKER_RUN_DSC_IMAGE_TAG=ds-dsc-run
DOCKER_RUN_DSA_IMAGE_TAG=ds-dsa-run
DOCKER_RUN_DSTEMPLATE_IMAGE_TAG=ds-dstemplate-run

K8S_WEB_URL=""

GitCount= $(shell git rev-list --count HEAD)
GitSHA1= $(shell git log -n1 --format=format:"%H")
branch=$(shell git symbolic-ref --short HEAD)
Branch=$(subst /,_,$(branch))

.PHONY: build  buildimage test clean vet deployimage dockertest deployimagedesign gitversion assetsmd5 pb
gitversion:
	rm -rf common/civersion.go
	echo 'package common '  >>common/civersion.go
	echo 'const GitVersion = "${Branch}_${GitCount}_${GitSHA1}"' >>common/civersion.go
	echo 'const DSServerName = "${BACKEND_SERVERJOB_NAME}"'  >>common/civersion.go

assetsmd5:
	tool/gen_assets_md5.py ${SlotsConfigPath} ${SystemConfigPath}

build:
	make gitversion
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
			-f ./Dockerfile.buildbinary .
	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) > ${DOCKER_BUILD_IMAGE_TAG}_${Version}_${Branch}_${GitCount}.tar.gz

buildimage:
	make gitversion
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
			-f ./Dockerfile.buildbinary .

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_GAME_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile.run -
	docker tag $(DOCKER_RUN_GAME_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_GAME_IMAGE_TAG):latest

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_DSC_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile_DSC.run -
	docker tag $(DOCKER_RUN_DSC_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_DSC_IMAGE_TAG):latest

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_DSA_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile_DSA.run -
	docker tag $(DOCKER_RUN_DSA_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_DSA_IMAGE_TAG):latest

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_DSTEMPLATE_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile_DSTemplate.run -
	docker tag $(DOCKER_RUN_DSTEMPLATE_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_DSTEMPLATE_IMAGE_TAG):latest


push_webserver:
	echo "Tagging docker image $(DOCKER_RUN_GAME_IMAGE_TAG) to $(K8S_WEB_URL)"
	docker tag $(DOCKER_RUN_GAME_IMAGE_TAG):${Version}_${Branch}_${GitCount} \
		$(K8S_WEB_URL):${Version}_${Branch}_${GitCount}
	echo "Start pushing"
	docker push $(K8S_WEB_URL):${Version}_${Branch}_${GitCount}


test: export DS_LOGLEVEL=info
test:
	go test  -cover -race -parallel 1 -p 1 -count=1 -v ./... | sed '/PASS/s//$(shell printf "\033[32mPASS\033[0m")/' | sed '/FAIL/s//$(shell printf "\033[31mFAIL\033[0m")/' | sed '/coverage/s//$(shell printf "\033[32mcoverage\033[0m")/'

# Redis 集群地址（与 docker-compose.infra.yaml / Dockerfile.test 一致）
REDIS_CLUSTER_URI=redis-node-1:7001,redis-node-2:7002,redis-node-3:7003,redis-node-4:7004,redis-node-5:7005,redis-node-6:7006

# 启动 Redis 集群 + MongoDB，并执行集群初始化：make infra / make infra-down
infra:
	docker compose -f docker-compose.infra.yaml up -d --wait
	docker compose -f docker-compose.infra.yaml --profile init run --rm redis-cluster-init

infra-down:
	docker compose -f docker-compose.infra.yaml --profile init down

# 构建测试镜像，挂到 ds-infra 网络跑单测（连 Redis 集群）
unit_test:
	$(MAKE) infra
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG)-test \
		-f ./Dockerfile.test .
	docker run --rm --network ds-infra \
		-e DS_CONFIGPATH=/src \
		-e DS_REDISURI=$(REDIS_CLUSTER_URI) \
		-e DS_MONGOURI=mongodb://mongodb-server:27017 \
		-e DS_MONGODBNAME=ds_test \
		-e DS_LOGLEVEL=info \
		$(DOCKER_BUILD_IMAGE_TAG)-test | tee unit_test.log

vet :
	go vet ./...

deployimage:
	make buildimage
	docker compose down
	docker compose up -d --build
pb:
	cd tools &&sed -i "s/\r//" gen_proto.sh &&chmod +x gen_proto.sh && ./gen_proto.sh
	
clean:
